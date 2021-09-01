package rbac

import (
	"github.com/stackrox/rox/generated/storage"
	v1 "k8s.io/api/rbac/v1"
)

type namespacedBindingID struct {
	namespace string
	uid       string
}

type namespacedBinding struct {
	roleRef  namespacedRoleRef   // The role that the subjects are bound to.
	subjects []namespacedSubject // The subjects that are bound to the referenced role.
}

func (b *namespacedBindingID) IsClusterBinding() bool {
	return len(b.namespace) == 0
}

func roleBindingToNamespacedBindingID(roleBinding *v1.RoleBinding) namespacedBindingID {
	return namespacedBindingID{namespace: roleBinding.GetNamespace(), uid: string(roleBinding.GetUID())}
}

func clusterRoleBindingToNamespacedBindingID(clusterRoleBinding *v1.ClusterRoleBinding) namespacedBindingID {
	return namespacedBindingID{namespace: "", uid: string(clusterRoleBinding.GetUID())}
}

func roleBindingToNamespacedBinding(roleBinding *v1.RoleBinding) *namespacedBinding {
	subjects := make([]namespacedSubject, 0, len(roleBinding.Subjects))
	for _, s := range getSubjects(roleBinding.Subjects) {
		// We only need this information for evaluating Deployment permission level,
		// so we can keep only ServiceAccount subjects (Pods cannot run as User or Group).
		if s.Kind == storage.SubjectKind_SERVICE_ACCOUNT {
			subjects = append(subjects, nsSubjectFromSubject(s))
		}
	}
	return &namespacedBinding{
		subjects: subjects,
		roleRef:  roleBindingToNamespacedRoleRef(roleBinding),
	}
}

func clusterRoleBindingToNamespacedBinding(clusterRoleBinding *v1.ClusterRoleBinding) *namespacedBinding {
	subjects := make([]namespacedSubject, 0, len(clusterRoleBinding.Subjects))
	for _, s := range getSubjects(clusterRoleBinding.Subjects) {
		// We only need this information for evaluating Deployment permission level,
		// so we can keep only ServiceAccount subjects (Pods cannot run as User or Group).
		if s.Kind == storage.SubjectKind_SERVICE_ACCOUNT {
			subjects = append(subjects, nsSubjectFromSubject(s))
		}
	}
	return &namespacedBinding{
		subjects: subjects,
		roleRef:  clusterRoleBindingToNamespacedRoleRef(clusterRoleBinding),
	}
}