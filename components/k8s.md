# Kubernetes — Internals & Ecosystem

Control plane, data plane, CNIs, service meshes, operators, GitOps, policy,
observability, CRDs. Component names here mirror upstream Kubernetes naming
(`kube-apiserver`, not "k8s api"), with aliases for familiar terms.

## Control plane

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.kube_apiserver` | kube-apiserver | control | A | REST API + auth + admission + watch server. Horizontally scalable. | api server, kube-apiserver | 3 HA apiservers behind HAProxy |
| `k8s.etcd` | etcd | control | A | Raft-consensus KV cluster behind apiserver (3–5 members). | etcd | 5-node etcd on NVMe |
| `k8s.kube_scheduler` | kube-scheduler | control | A | Binds pending Pods to Nodes via predicates + priorities (profiles). | scheduler | Custom scheduler for GPU |
| `k8s.kube_controller_manager` | kube-controller-manager | control | A | Runs 40+ controllers (Deployment, ReplicaSet, Node, Endpoint, Service-Account…). | controller manager | Bootstrap cluster CM |
| `k8s.cloud_controller_manager` | cloud-controller-manager | control | A | Provider-specific bits (Node, Route, Service LB) extracted from kube-cm. | ccm, cloud controller | AWS LB CCM |
| `k8s.admission_mutating` | Mutating Admission | control | A | Webhook mutates objects (sidecar injection, defaulting). | mutating webhook | Sidecar injector |
| `k8s.admission_validating` | Validating Admission | control | A | Webhook admits / denies objects; Kyverno, Gatekeeper. | validating webhook | Policy webhook |
| `k8s.audit_policy` | Audit Policy | control | A | Structured audit log of all API requests. | audit log, k8s audit | File audit → Loki |

## Data plane (node)

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.kubelet` | kubelet | node | A | Per-node agent; pod lifecycle via CRI, CSI, CNI. | kubelet | Cordon + drain kubelet |
| `k8s.kube_proxy` | kube-proxy | node | A | iptables/ipvs/nftables forwarder for Services. | kube-proxy, kubeproxy | IPVS mode for 10k services |
| `k8s.container_runtime_containerd` | containerd | node | I | Default runtime used by EKS / GKE / AKS. | containerd, cri | containerd 1.7 on nodes |
| `k8s.container_runtime_crio` | CRI-O | node | A | CRI runtime used by OpenShift. | crio, cri-o | OpenShift node |
| `k8s.runtime_gvisor` | gVisor | node | A | User-space sandbox runtime (Google runsc). | gvisor | Multi-tenant pods |
| `k8s.runtime_kata` | Kata | node | A | VM-sandboxed pods via QEMU/Firecracker/Cloud Hypervisor. | kata containers | VM-isolated tenant |

## Add-ons

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.coredns` | CoreDNS | dns | B | In-cluster DNS (service discovery, pod DNS). | coredns, cluster dns | DNS for services |
| `k8s.metrics_server` | Metrics Server | observability | I | `kubectl top` + HPA source. | metrics server | HPA metrics |
| `k8s.dashboard` | Dashboard | tooling | B | Web UI (k8s.io/dashboard). | kubernetes dashboard | Ops UI |
| `k8s.external_dns` | external-dns | dns | I | Sync Services/Ingresses to Route53 / Cloud DNS / CF. | external-dns | DNS automation |
| `k8s.cert_manager` | cert-manager | crypto | I | Issue + renew TLS via ACME, Vault, CA. | cert-manager | Let's Encrypt certs |

## Workloads

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.pod` | Pod | workload | B | Unit of deployment; 1+ containers sharing net + IPC + storage. | pod | Sidecar + main container |
| `k8s.deployment` | Deployment | workload | B | Rolling-update abstraction on top of ReplicaSet. | deployment, deploy | Stateless app |
| `k8s.replicaset` | ReplicaSet | workload | I | Maintains N replicas of a Pod template. | replicaset, rs | Managed by Deployment |
| `k8s.statefulset` | StatefulSet | workload | I | Stable identity + ordered bootstrap for stateful apps. | statefulset, sts | Kafka brokers |
| `k8s.daemonset` | DaemonSet | workload | I | One Pod per Node (or per match). | daemonset, ds | Fluent Bit on every node |
| `k8s.job` | Job | workload | I | Run-to-completion batch. | job | One-shot migration |
| `k8s.cronjob` | CronJob | workload | I | Time-scheduled Jobs. | cronjob | Nightly report |
| `k8s.init_container` | Init Container | workload | I | Ordered setup containers before main containers. | init container | Wait-for-db init |
| `k8s.sidecar_container` | Sidecar | workload | A | Long-running helper (proxy, logger, mesh injector). | sidecar | Envoy sidecar |
| `k8s.ephemeral_container` | Ephemeral Container | workload | A | Ad-hoc debug container attached to a running Pod (`kubectl debug`). | ephemeral container | Debug prod pod |

## Config & storage

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.configmap` | ConfigMap | config | B | Non-secret key/value config; mount as env/file. | configmap, cm | App config |
| `k8s.secret` | Secret | config | B | base64 opaque; SealedSecret / ExternalSecret for GitOps. | secret | DB credentials |
| `k8s.pvc` | PersistentVolumeClaim | storage | I | Request for storage satisfied by StorageClass. | pvc, persistent volume claim | Postgres PVC |
| `k8s.pv` | PersistentVolume | storage | I | Cluster-level volume resource (dynamic or static). | pv, persistent volume | EBS PV |
| `k8s.storageclass` | StorageClass | storage | I | Provisioner + params for dynamic PVs. | storage class, sc | gp3-encrypted SC |
| `k8s.csi_driver` | CSI Driver | storage | A | Container Storage Interface plugin (EBS, CSI-Ceph, Portworx…). | csi | aws-ebs-csi |
| `k8s.volumesnapshot` | VolumeSnapshot | storage | A | Point-in-time snapshot object bound to CSI. | volume snapshot | EBS snapshot via CSI |

## Networking

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.service_clusterip` | Service (ClusterIP) | network | B | Virtual cluster IP + port; default Service type. | service, clusterip | Microservice reachability |
| `k8s.service_nodeport` | Service (NodePort) | network | I | Exposes service on each Node's IP:port. | nodeport | Dev access |
| `k8s.service_loadbalancer` | Service (LoadBalancer) | network | B | Cloud LB wired to a Service. | loadbalancer service, type lb | Cloud LB per service |
| `k8s.service_externalname` | Service (ExternalName) | network | I | DNS CNAME to external host. | externalname service | DNS alias |
| `k8s.service_headless` | Headless Service | network | A | `clusterIP: None` — returns pod DNS records directly. | headless service | StatefulSet DNS |
| `k8s.ingress` | Ingress | network | B | L7 routing manifest; needs an ingress controller. | ingress | Nginx ingress |
| `k8s.gateway_api` | Gateway API | network | A | Successor to Ingress — Gateway, HTTPRoute, GRPCRoute, TLSRoute, etc. | gateway api | Gateway API rollout |
| `k8s.networkpolicy` | NetworkPolicy | network | A | L3/L4 pod firewall (needs CNI support). | network policy | Default-deny ns |
| `k8s.endpointslice` | EndpointSlice | network | A | Scalable replacement for Endpoints (partitioned). | endpoint slice | Large-service discovery |
| `k8s.cni_calico` | Calico | network | A | CNI + NetworkPolicy engine; BGP or IPIP or VXLAN. | calico | CNI choice |
| `k8s.cni_cilium` | Cilium | network | A | eBPF-based CNI + mesh + network policy + observability. | cilium, ebpf cni | Cilium Service Mesh |
| `k8s.cni_flannel` | Flannel | network | I | Simple overlay CNI (VXLAN). | flannel | Small clusters |
| `k8s.cni_weave` | Weave | network | E | Mesh CNI (⚠ less maintained). | weave | Legacy |
| `k8s.cni_aws_vpc` | AWS VPC CNI | network | A | Pods get real VPC IPs via ENI prefix delegation. | aws vpc cni, amazon vpc cni | EKS default |
| `k8s.cni_azure_cni` | Azure CNI | network | A | Pods on VNet with CNI-overlay / no-overlay modes. | azure cni | AKS default |
| `k8s.cni_gke` | GKE CNI | network | A | Netd + Dataplane V1 / V2 (Cilium). | gke cni | GKE default |
| `k8s.multus` | Multus | network | E | Multiple NICs per Pod (telco use-cases). | multus | SR-IOV + macvlan |

## Ingress controllers & meshes

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.ingress_nginx` | ingress-nginx | network | I | NGINX-based ingress controller. | nginx ingress | Default ingress |
| `k8s.ingress_traefik` | Traefik | network | I | Traefik proxy + ingress controller + Gateway API. | traefik | Edge proxy |
| `k8s.ingress_haproxy` | HAProxy Ingress | network | A | HAProxy-based ingress. | haproxy ingress | HAProxy L7 |
| `k8s.ingress_contour` | Contour | network | A | Envoy-based Ingress + HTTPProxy CRD. | contour | Envoy ingress |
| `k8s.ingress_kong` | Kong Ingress | network | A | Kong Gateway as ingress controller. | kong | API gateway ingress |
| `k8s.service_mesh_istio` | Istio | mesh | A | Envoy sidecars (or ambient), CRDs for VirtualService, DestinationRule, ServiceEntry. | istio | mTLS + canary |
| `k8s.service_mesh_linkerd` | Linkerd | mesh | A | Rust-micro-proxy mesh; lean + fast. | linkerd | mTLS default |
| `k8s.service_mesh_cilium` | Cilium Service Mesh | mesh | A | eBPF-first mesh, sidecarless. | cilium mesh | eBPF mesh |
| `k8s.service_mesh_consul` | Consul Connect | mesh | A | HashiCorp mesh + service discovery. | consul | Multi-cluster mesh |
| `k8s.service_mesh_kuma` | Kuma | mesh | A | Kong's Envoy-based mesh. | kuma | Multi-zone mesh |
| `k8s.service_mesh_osm` | Open Service Mesh | mesh | E | CNCF Envoy mesh (⚠ retired). | osm | Legacy |
| `k8s.istio_ambient` | Istio Ambient | mesh | A | Sidecarless Istio via ztunnel + waypoint. | istio ambient | No-sidecar mode |

## Autoscaling

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.hpa` | HPA | autoscale | I | Horizontal Pod Autoscaler v2 on CPU/memory/custom metrics. | hpa, horizontal pod autoscaler | Scale API on RPS |
| `k8s.vpa` | VPA | autoscale | A | Vertical Pod Autoscaler — recommend/auto resize pod resources. | vpa, vertical pod autoscaler | Right-size cron |
| `k8s.cluster_autoscaler` | Cluster Autoscaler | autoscale | I | Resize node groups up/down on unschedulable pods. | cluster autoscaler, ca | Scale EKS NG |
| `k8s.karpenter` | Karpenter | autoscale | A | Faster + smarter node autoscaler; NodePool CRD. | karpenter | Mix Graviton + Spot |
| `k8s.keda` | KEDA | autoscale | A | Event-driven autoscaling (Kafka lag, SQS depth, Prom). | keda | Scale on queue depth |

## Security / policy

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.rbac_role` | Role / RoleBinding | security | I | Namespace-scoped permissions. | rbac, role binding | App-specific role |
| `k8s.rbac_clusterrole` | ClusterRole / Binding | security | I | Cluster-wide RBAC. | clusterrole | Read-only all ns |
| `k8s.serviceaccount` | ServiceAccount | security | I | Identity for pods; backs cluster tokens or OIDC. | serviceaccount, sa | Pod identity |
| `k8s.pod_security_standards` | Pod Security Standards | security | A | Baseline / Restricted / Privileged profiles (successor to PSP). | pss, pod security | Restricted namespace |
| `k8s.psp` | PodSecurityPolicy | security | E | ⚠ Removed in 1.25 — use PSS/OPA/Kyverno. | psp | Legacy |
| `k8s.opa_gatekeeper` | OPA Gatekeeper | security | A | Policy engine w/ Rego (Open Policy Agent). | gatekeeper, opa | CIS policies |
| `k8s.kyverno` | Kyverno | security | A | Kubernetes-native policy (YAML). | kyverno | Mutate + validate |
| `k8s.kubewarden` | Kubewarden | security | A | WASM-powered admission policies. | kubewarden | WASM policies |
| `k8s.external_secrets` | External Secrets | security | A | Sync secrets from AWS SM / Vault / Key Vault / GCP SM. | external secrets, eso | Vault → Secret |
| `k8s.sealed_secrets` | Sealed Secrets | security | A | Encrypt secrets for Git (Bitnami). | sealed secrets | GitOps secrets |
| `k8s.sops` | SOPS | security | A | Mozilla SOPS with KMS/Age encrypt Git-stored secrets. | sops | Age-encrypted files |

## GitOps & delivery

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.argocd` | Argo CD | cd | I | GitOps controller reconciling repo → cluster. | argocd, argo cd | App-of-apps pattern |
| `k8s.argo_rollouts` | Argo Rollouts | cd | A | Canary + blue-green w/ analysis templates. | argo rollouts | Canary with Prom analysis |
| `k8s.argo_workflows` | Argo Workflows | orchestration | A | K8s-native DAG workflow engine. | argo workflows | ML pipeline |
| `k8s.argo_events` | Argo Events | events | A | Event triggers for workflows / GitOps. | argo events | Event-driven CI |
| `k8s.flux` | Flux CD | cd | I | GitOps toolkit (source/kustomize/helm/notification controllers). | flux, fluxcd | GitOps for clusters |
| `k8s.helm` | Helm | packaging | B | Package manager; templates + releases. | helm | Prom stack install |
| `k8s.kustomize` | Kustomize | packaging | I | Overlay-based YAML templating. | kustomize | Env overlays |
| `k8s.tekton` | Tekton | ci | A | K8s-native CI (TaskRun, PipelineRun). | tekton | Cloud-native CI |
| `k8s.jenkins_x` | Jenkins X | cd | E | GitOps + Tekton (⚠ niche). | jenkins x | Legacy |
| `k8s.knative_serving` | Knative Serving | serverless | A | Scale-to-zero HTTP serving on K8s. | knative serving, knative | Request-driven |
| `k8s.knative_eventing` | Knative Eventing | events | A | CloudEvents broker/trigger on K8s. | knative eventing | Event mesh |
| `k8s.crossplane` | Crossplane | iac | A | Provision cloud infra via K8s CRDs (managed resources). | crossplane | TF-as-CRDs |
| `k8s.capi` | Cluster API | infra | A | Declarative cluster lifecycle (Machine, MachineDeployment…). | capi, cluster api | Multi-cluster provisioning |
| `k8s.operator_sdk` | Operator SDK | platform | A | Framework to build custom controllers (Go/Helm/Ansible). | operator-sdk, operator pattern | DB operator |
| `k8s.crd` | CRD | platform | A | CustomResourceDefinition extending the API. | crd, custom resource | Org-specific API |

## Observability

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `k8s.prometheus` | Prometheus | observability | I | Pull metrics w/ PromQL + alerting. | prometheus, prom | Service SLOs |
| `k8s.grafana` | Grafana | observability | I | Dashboards over metrics/logs/traces. | grafana | Exec dashboards |
| `k8s.loki` | Loki | observability | A | Log store indexed by labels, LogQL. | loki | Logs w/ labels |
| `k8s.tempo` | Tempo | observability | A | Trace backend w/ TraceQL. | tempo | Trace store |
| `k8s.mimir` | Mimir | observability | A | Horizontally scalable Prometheus TSDB (Grafana Labs). | mimir | Scaled Prom TSDB |
| `k8s.thanos` | Thanos | observability | A | Global query + long-term Prom storage (object store). | thanos | LTS metrics |
| `k8s.victoria_metrics` | Victoria Metrics | observability | A | High-perf Prom-compatible TSDB. | victoria metrics, vm | TS at scale |
| `k8s.jaeger` | Jaeger | observability | I | Distributed tracing backend (CNCF). | jaeger | Trace explorer |
| `k8s.open_telemetry` | OpenTelemetry | observability | I | Unified metrics / logs / traces SDK + collector. | otel, open telemetry | OTLP pipeline |
| `k8s.fluentbit` | Fluent Bit | observability | I | Log shipper sidecar / node agent. | fluent bit | Node log ship |
| `k8s.fluentd` | Fluentd | observability | I | Log processor + router. | fluentd | Legacy log agent |
| `k8s.falco` | Falco | security | A | Runtime security (syscalls + K8s audit). | falco | Runtime threat |
| `k8s.trivy` | Trivy | security | I | Image + IaC + SBOM scanner. | trivy | Image vulns |

---

**Entry count: ~65.** Covers K8s upstream + CNCF graduated/incubating ecosystem.
