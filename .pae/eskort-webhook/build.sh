rm -fr eskort-webhook-1.0.0.tgz
rm -fr ./eskort-webhook
helm del --purge kubebuilder
cp -r chart eskort-webhook
helm package ./eskort-webhook
helm install ./eskort-webhook --name kubebuilder --namespace eskort

