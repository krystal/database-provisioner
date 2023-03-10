package v1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var mysqldatabaselog = logf.Log.WithName("mysqldatabase-resource")

func (r *MySQLDatabase) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-databases-k8s-k-io-v1-mysqldatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=databases.k8s.k.io,resources=mysqldatabases,verbs=create;update,versions=v1,name=mmysqldatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &MySQLDatabase{}

func (r *MySQLDatabase) Default() {
	mysqldatabaselog.Info("default", "name", r.Name)

	if r.Spec.ServerName == "" {
		r.Spec.ServerName = "default"
	}

	if r.Spec.ConnectionDetailsSecretName == "" {
		r.Spec.ConnectionDetailsSecretName = fmt.Sprintf("%s-database-details", r.Name)
	}

	r.Status.Created = false
}

//+kubebuilder:webhook:path=/validate-databases-k8s-k-io-v1-mysqldatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=databases.k8s.k.io,resources=mysqldatabases,verbs=create;update,versions=v1,name=vmysqldatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &MySQLDatabase{}

func (r *MySQLDatabase) ValidateCreate() error {
	mysqldatabaselog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MySQLDatabase) ValidateUpdate(old runtime.Object) error {
	mysqldatabaselog.Info("validate update", "name", r.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MySQLDatabase) ValidateDelete() error {
	mysqldatabaselog.Info("validate delete", "name", r.Name)

	return nil
}
