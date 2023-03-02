package controller

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jimeh/rands"
	databasesv1 "github.com/krystal/k8s-database-provisioner/api/v1"
)

type MySQLDatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=databases.k8s.k.io,resources=mysqldatabases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=databases.k8s.k.io,resources=mysqldatabases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=databases.k8s.k.io,resources=mysqldatabases/finalizers,verbs=update
//+kubebuilder:rbac:groups=databases.k8s.k.io,resources=mysqlservers,verbs=list;get;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;delete

const (
	ownerAnnotationKey = "databases.k8s.io/owner"
	finalizerName      = "databases.k8s.io/finalizer"
)

func (r *MySQLDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var database databasesv1.MySQLDatabase
	if err := r.Get(ctx, req.NamespacedName, &database); err != nil {
		log.Error(err, "unabled to fetch MySQLDatabase object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("found database")

	// add a finalizer immediaetly if missing
	if !controllerutil.ContainsFinalizer(&database, finalizerName) {
		log.Info("adding finalizer to database", "finalizer", finalizerName)
		controllerutil.AddFinalizer(&database, finalizerName)
		if err := r.Update(ctx, &database); err != nil {
			r.logErrorAndUpdate(ctx, log, database, err, "could not add finalizer")
			return ctrl.Result{}, err
		}
	}

	// find the server that is referenced by the database in the ServerName field.
	var server databasesv1.MySQLServer
	if err := r.Get(ctx, client.ObjectKey{Namespace: "", Name: database.Spec.ServerName}, &server); err != nil {
		r.logErrorAndUpdate(ctx, log, database, err, "did not find MySQLServer with given name")
		return ctrl.Result{}, err
	}

	log.Info("found server to host database", "server", server.ObjectMeta.Name)

	// check to see if we already have any connection details for this database, if we do we want to use
	// the details from there for determining what to use when provisioning the database
	var connectionDetailsSecret v1.Secret
	newSecret := false
	var databaseNameWithNamespace string
	var databasePassword string
	err := r.Get(ctx, client.ObjectKey{Namespace: database.ObjectMeta.Namespace, Name: database.Spec.ConnectionDetailsSecretName}, &connectionDetailsSecret)
	if errors.IsNotFound(err) {
		log.Info("no existing secret found", "secretName", database.Spec.ConnectionDetailsSecretName)
		newSecret = true
		connectionDetailsSecret = v1.Secret{
			ObjectMeta: ctrl.ObjectMeta{
				Name:      database.Spec.ConnectionDetailsSecretName,
				Namespace: database.ObjectMeta.Namespace,
				Annotations: map[string]string{
					ownerAnnotationKey: fmt.Sprintf("%s/%s", database.ObjectMeta.Namespace, database.ObjectMeta.Name),
				},
			},
		}

		// work out what the database should be called
		databaseNameWithNamespace = fmt.Sprintf("%s_%s", database.ObjectMeta.Namespace, database.ObjectMeta.Name)

		// generate a random string with 24 charaters for the password
		databasePassword, err = rands.Alphanumeric(24)
		if err != nil {
			r.logErrorAndUpdate(ctx, log, database, err, "unable to generate random password")
			return ctrl.Result{}, err
		}
		log.Info("auto generated new password for database")
	} else if err != nil {
		r.logErrorAndUpdate(ctx, log, database, err, "unable to get connection details secret")
		return ctrl.Result{}, err
	} else {
		log.Info("got existing secret", "details", connectionDetailsSecret.Data)
		databaseNameWithNamespace = string(connectionDetailsSecret.Data["databaseName"])
		databasePassword = string(connectionDetailsSecret.Data["password"])
	}

	log.Info("determined database name", "databaseNameWithNamespace", databaseNameWithNamespace)

	// return error if the database name contains anything other than letters, numbers, hyphens and underscores
	databaseNameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !databaseNameRegex.MatchString(databaseNameWithNamespace) {
		r.logErrorAndUpdate(ctx, log, database, err, "database name contains invalid characters")
		return ctrl.Result{}, err
	}

	// return an error if the database name is longer than 64 characters
	if len(databaseNameWithNamespace) > 64 {
		r.logErrorAndUpdate(ctx, log, database, err, "database name is too long")
		return ctrl.Result{}, err
	}

	// connect to the host server
	log.Info("connecting to mysql server", "host", server.Spec.Host, "port", server.Spec.Port, "username", server.Spec.Username)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", server.Spec.Username, server.Spec.Password, server.Spec.Host, server.Spec.Port))
	if err != nil {
		r.logErrorAndUpdate(ctx, log, database, err, "unable to connect to MySQL server")
		return ctrl.Result{}, err
	}
	defer db.Close()

	log.Info("connected to mysql server", "server", server.Spec.Host, "username", server.Spec.Username)

	if !database.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("database has been deleted")
		// delete the database from the backend - scary, scary stuff
		row := db.QueryRow("drop database `" + databaseNameWithNamespace + "`")
		if row.Err() != nil {
			r.logErrorAndUpdate(ctx, log, database, row.Err(), "unable to delete database")
			return ctrl.Result{}, row.Err()
		}
		log.Info("deleted database", "database", databaseNameWithNamespace)

		log.Info("removing finalizer")
		controllerutil.RemoveFinalizer(&database, finalizerName)
		if err := r.Update(ctx, &database); err != nil {
			r.logErrorAndUpdate(ctx, log, database, err, "could not remove finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// go and see if the database already exists, if it does we don't need to create it
	var existingDatabaseName string
	err = db.QueryRow("select schema_name from information_schema.schemata where schema_name = ?", databaseNameWithNamespace).Scan(&existingDatabaseName)
	if err != nil {
		log.Info("Error querying for existing database", "database", databaseNameWithNamespace)
	}

	// create the database if it doesn't already exist
	if existingDatabaseName == "" {
		row := db.QueryRow("create database `" + databaseNameWithNamespace + "`")
		if row.Err() != nil {
			r.logErrorAndUpdate(ctx, log, database, row.Err(), "unable to create database")
			return ctrl.Result{}, row.Err()
		}
		log.Info("created database", "database", databaseNameWithNamespace)
	} else {
		log.Info("database already exists", "database", databaseNameWithNamespace)
	}

	// verify access has been set up properly to grant access to the database
	row := db.QueryRow(fmt.Sprintf("grant all on `%s`.* to `%s`@`%%` identified by \"%s\"", databaseNameWithNamespace, databaseNameWithNamespace, databasePassword))
	if row.Err() != nil {
		r.logErrorAndUpdate(ctx, log, database, row.Err(), "unable to grant access to database")
		return ctrl.Result{}, row.Err()
	}
	log.Info("set grant", "database", existingDatabaseName)

	// ensure the secret is up to date
	connectionDetailsSecret.StringData = make(map[string]string)
	connectionDetailsSecret.StringData["databaseName"] = databaseNameWithNamespace
	connectionDetailsSecret.StringData["username"] = databaseNameWithNamespace
	connectionDetailsSecret.StringData["password"] = databasePassword
	connectionDetailsSecret.StringData["host"] = server.Spec.Host

	if newSecret {
		err = r.Create(ctx, &connectionDetailsSecret)
		if err != nil {
			r.logErrorAndUpdate(ctx, log, database, err, "unable to create connection details secret")
			return ctrl.Result{}, err
		}
	} else {
		err = r.Update(ctx, &connectionDetailsSecret)
		if err != nil {
			r.logErrorAndUpdate(ctx, log, database, err, "unable to update connection details secret")
			return ctrl.Result{}, err
		}
	}

	// update the database status record
	database.Status.Created = true
	database.Status.Error = ""
	err = r.Status().Update(ctx, &database)
	if err != nil {
		r.logErrorAndUpdate(ctx, log, database, err, "unable to update database status")
		return ctrl.Result{}, err
	}

	log.Info("reconcile complete")

	return ctrl.Result{}, nil
}

func (r *MySQLDatabaseReconciler) logErrorAndUpdate(ctx context.Context, log logr.Logger, database databasesv1.MySQLDatabase, err error, message string) error {
	log.Error(err, message)
	database.Status.Error = message
	return r.Status().Update(ctx, &database)
}

func (r *MySQLDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasesv1.MySQLDatabase{}).
		Complete(r)
}
