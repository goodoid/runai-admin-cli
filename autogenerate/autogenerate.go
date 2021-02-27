// THIS FILE IS AUTO GENERATED ON MAKE COMMAND - DO NOT EDIT
 
package autogenerate 
var PreInstallYaml = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: researcher-service
  namespace: runai
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: researcher-service
rules:
  - apiGroups:
      - ""
      - apps
      - run.ai
    resources:
      - configmaps
      - namespaces
      - pods
      - pods/log
      - projects
      - runaijobs
      - statefulsets
    verbs:
      - create
      - delete
      - get
      - list
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: researcher-service
subjects:
  - kind: ServiceAccount
    name: researcher-service
    namespace: runai
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: researcher-service
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: runai-scheduler
  namespace: runai
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: runai-scheduler-ro
rules:
  - apiGroups:
      - ""
      - batch
      - policy
      - run.ai
      - scheduling.k8s.io
      - scheduling.incubator.k8s.io
      - storage.k8s.io
    resources:
      - departments
      - jobs
      - nodes
      - persistentvolumes
      - persistentvolumeclaims
      - poddisruptionbudgets
      - podgroups
      - pods
      - priorityclasses
      - queues
      - runaijobs
      - storageclasses
    verbs:
      - get
      - list
      - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: runai-scheduler-ro
subjects:
  - kind: ServiceAccount
    name: runai-scheduler
    namespace: runai
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: runai-scheduler-ro
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: runai-scheduler-rw
rules:
  - apiGroups:
      - ""
      - apps
      - run.ai
      - scheduling.incubator.k8s.io
    resources:
      - configmaps
      - events
      - persistentvolumeclaims
      - podgroups
      - pods
      - pods/status
      - pods/binding
      - runaijobs
      - statefulsets
    verbs:
      - create
      - get
      - list
      - patch
      - update
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: runai-scheduler-rw
subjects:
  - kind: ServiceAccount
    name: runai-scheduler
    namespace: runai
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: runai-scheduler-rw
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: runai-db
  namespace: runai
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: runai-db-migrations
  namespace: runai
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: runai-vgpu
  namespace: runai
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: runai-nvidia-device-plugin
  namespace: runai
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: runai-nvidia-device-plugin
rules:
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - get
      - list
      - update
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
      - list
      - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: runai-nvidia-device-plugin
subjects:
  - kind: ServiceAccount
    name: runai-nvidia-device-plugin
    namespace: runai
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: runai-nvidia-device-plugin
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: runai-agent
  namespace: runai
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: runai-agent
rules:
  - apiGroups:
      - run.ai
    resources:
      - projects
    verbs:
      - create
      - delete
      - get
      - list
      - update
  - apiGroups:
      - scheduling.incubator.k8s.io
    resources:
      - departments
    verbs:
      - create
      - delete
      - get
      - list
      - update
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - clusterroles
    resourceNames:
      - runai-cli-index-map-editor
    verbs:
      - bind
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - clusterroles
    resourceNames:
      - runai-job-viewer
    verbs:
      - bind
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - rolebindings
    resourceNames:
      - runai-cli-index-map-editor
    verbs:
      - get
      - list
      - update
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - clusterrolebindings
    resourceNames:
      - runai-job-viewer
    verbs:
      - get
      - list
      - update
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: runai-agent
subjects:
  - kind: ServiceAccount
    name: runai-agent
    namespace: runai
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: runai-agent
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: runaiconfigs.run.ai
spec:
  group: run.ai
  names:
    kind: RunaiConfig
    listKind: RunaiConfigList
    plural: runaiconfigs
    singular: runaiconfig
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      type: object
      x-kubernetes-preserve-unknown-fields: true
  versions:
    - name: v1
      served: true
      storage: true
---
apiVersion: v1
kind: Namespace
metadata:
  name: runai
---
apiVersion: v1
kind: Secret
metadata:
  name: gcr-secret
  namespace: runai
data:
  .dockerconfigjson: eyAiYXV0aHMiOiB7ICJnY3IuaW8iOiB7ICJhdXRoIjogIlgycHpiMjVmYTJWNU9uc2dJQ0owZVhCbElqb2dJbk5sY25acFkyVmZZV05qYjNWdWRDSXNJQ0FpY0hKdmFtVmpkRjlwWkNJNklDSnlkVzR0WVdrdGNISnZaQ0lzSUNBaWNISnBkbUYwWlY5clpYbGZhV1FpT2lBaVlXRXhabVpqTXpsaVpXUXhaamhqTURRMU1tSTVOMkV4T1RrMVptUmxObVprTXpNMk1tUTFNQ0lzSUNBaWNISnBkbUYwWlY5clpYa2lPaUFpTFMwdExTMUNSVWRKVGlCUVVrbFdRVlJGSUV0RldTMHRMUzB0WEc1TlNVbEZkbEZKUWtGRVFVNUNaMnR4YUd0cFJ6bDNNRUpCVVVWR1FVRlRRMEpMWTNkbloxTnFRV2RGUVVGdlNVSkJVVVJLZDNGQ1NFVmliVll2UTAxNVhHNTFlWEYxYlVKcVZtOHlNRUpWYzJaUGFEWjFkRXRDVGpoUVVrUTBhRmt6Wm5Wek5rbHJaelZEYmtrMVVVZDJRbUU1TlU1RGJrcGhhSFZKVkZGTVYwZE1YRzVFZEdaMk9XOTNWVEZSYm5STGMwRnZPVnBJYm1jdlFWTkZla1JsUkdadVFrVndWbm9yU0ROR056SXlVemhIUlVGcGRXVnFlVGhsZWtWUlZEUnZZekpMWEc1V1V6Wm5WbW95T0ZZNVQyOXVaek5sVDJGMGJtUklRbFZxUmxodFFXRk9kbUpzUWtrMFRFWlZkVXhvUm1WbVkzbERMMEZvVERKYVF6bGtjRlJaVUhsYVhHNXFkRWxJVUhGVVR6QnlkSGhRVEdrNGRtaDBRV2gyTVUxb1FuWkViVEpFUmpCWGIwVTRTMlJFSzI5NldWUkxUVkJ6U1hCclJHRm1Nbk00Y0Zob1lXcE5YRzV3YUVkRFFsYzNhekE0U25aeVNpdElaMVk1YlRFNFMzaG1NU3R1YVVkcWJrMWhSV0pOVFRjM2MyZFRXaTkxUTNoM01rVXdhbmR1ZGpGWlMzZFZjQ3RtWEc1a1pGTnhRbmh6YmtGblRVSkJRVVZEWjJkRlFWaGlLMnh2WlVKSGFqYzJZM3BHU0ZSMk1WUkdSbFl2Wlc5eFRFbFlUbk5HYkcxMmMzZGxiazlJVDNoU1hHNTRRV1ZXVDBSb1JtdEJXVmQ1Ym00MUwxRnlWWHBxY21Oak1FTTVNV015UVdGYVJEUklXRkVyUjNRdmVYZE9WVUZtVVdVclUwRnJlVmd2VDNrelZITTFYRzVPWkVabmVVbGFTMU4wU3pVME0wUXlXV0kwY0dKQ05tZExSVUpqVVhaMFRDdHNPR2xVZWxwRVkxZFRUalJQWlhkSlpscERRWE40UkRsalZsQnNTazlUWEc0MlExZDVLMVZEYlZONmRYbDFhbms1Tlc5VVJscExTQzgxUWpaeVFXdzNOUzh5TDNORlVVa3hVaXRYY1ZOUmRVdDNjbmxGVG5SaU9DdExVazFrVGxaRVhHNXVaVkp1VVZWWmFqSjZLM1p2YmtsVGJqRm1hSGgyWjBvMGEzaHZlRWR1TUdOeWJERnBaMFpuY0RaYU16SkpRblZUVEZFNE9FSnlhVkJFVW14M016bGhYRzVaU1c5MVYxUXpTR1F3VmtadFpITlZSMnBTWjNKamVHOU5TekoxU0ZRd2NGbG9WekUyWW1aa2IxRkxRbWRSUkhScWEyRTVhR3c0U1dWeFdWQXZWak4yWEc1NlVUaFhkRzlEVVhWMk5sRkZMMGg2U1ZSNmEydGhZVm81YVRKVU56azVlSGd2WW05eVJsWjRlSEJrTkRBNU56TXhMMXB6ZUhGUFFYWmFTV05oU3pVMlhHNTBPRFJ5YTFnMk5VUkxkSGhYY2xkcmJuaDJNeXQ1UlRKRVIySkdTek5ZVFhSdU4zZE9halJ0UkhkTE0xSk9TVlZJUWpGSFRWbGpRV3hDZEdveFprODBYRzVaVkdwSGVURXljV0Z2VjJGTmFqWnBNRW96VWtSbVQybGhkMHRDWjFGRVdtSk9OVTgwUmxsdGFFRnVMM05rTm5oVlVESjNNRlZvTkZaWEwzRnJhbE0xWEc1d01VSnBhV2xsUTBkcWJtUTJSV1pDSzJsR0syVnpMMnhEYkRaR1ZtVkZaMWh4ZHpsRVdEbFFNemh4YW1SeFIyeFZiVWhQV0hwUlduTTBia0ZRTkc1c1hHNTJXRWxYYVdWVE1GQjViV3B4VUZWWU9GbEdZMlphUzNremNEazJaa2N4YUVWWFRtcElNVEI1TXpGUGNtZEJURll3VlhSRGFIRjVhSGRhWWs1eGRDODJYRzVvZUdWU2FsbEJlRTVSUzBKblJHWlZSMmhtYzJadFZWWjJabFpCUkVaWFVrVmFlVFkwTVdkblQycGtUMGRMZVZaQlUxTlBabTFNYzJ0cFYxVlhRMVEzWEc1Wk4wZHRNM0ZZVVd0RlUySk9iV3d3TTJKelQzRTVjRlJ6ZDFSMVRGTk5Na1V3VURFck5YZDBkVUpVTlhodWNXSXdaM3BxWm1ocFpuSllPV3hEVm1oR1hHNUJNbmRqY3pGd2NXRkxOemxuTkcxeFYyTTNibFZPZWpNNFlpOTVlR3RhYzNkMFZXeGphWFpqZDA4eFQwcGFOVW92VDNwdGNFZ3hiRUZ2UjBGUVRqRllYRzVHYVcxRGRFZEdLMFY0UVU0d1VVWnNTWGg0VXpsNWVXcHRZbmt6SzJOcE1tNW1PR2wwUXpjelUwRkhRVXBRVFV0alZXVmlOM1puUzBsaVZUWjNhamNyWEc1YVJrUnJPVTB3YmtGeU9YY3pUVVJHUmtkd1pWRldkV0pGYWtFelVVSk1ZVmg1YWxjeGQxcG1aRFp6VTJkV1RtWTNVelZTTlV4VGFGaEVOVGgxUWtkUlhHNURkVEUyZGpSVU5DOVRaRzF2T0doc1JYZG9NRkozZGxWV1ZuRnBVRXBYV1hGWGR6bEhTV3REWjFsRlFXaGhORFZuV1UxcmRtOU1kRVpPVVVOUE5HTnFYRzVUZVhWNlFtWmFiM2hHV2pWTmNETkJSVE4xZFRoRVRuazJlbkpPTlVwbVVDdDZjVmxIYkhsMFFVVm5ObU5CVjNCak9VZ3hjWG92UjBSc1JGTmtUbkZEWEc1R1VuTTNkMWw1V1dwMFIzTm9VV3BPYjFWMWVXNTZWelZaYldSMGFHaHZiMHh5WW1ScFR6UnZZa2hYYUZOblRHVm1WR0YyT0RKaWR6aHZlV3gwUkV0TlhHNXRNbFZqVTFZNE5WVlVkbFJ3TUVobldrMWlNUzgzTkQxY2JpMHRMUzB0UlU1RUlGQlNTVlpCVkVVZ1MwVlpMUzB0TFMxY2JpSXNJQ0FpWTJ4cFpXNTBYMlZ0WVdsc0lqb2dJbWRqY2kxd2RXeHNRSEoxYmkxaGFTMXdjbTlrTG1saGJTNW5jMlZ5ZG1salpXRmpZMjkxYm5RdVkyOXRJaXdnSUNKamJHbGxiblJmYVdRaU9pQWlNVEUxT0RFMU9EQTNPREF6TkRZeU16VXhNemd6SWl3Z0lDSmhkWFJvWDNWeWFTSTZJQ0pvZEhSd2N6b3ZMMkZqWTI5MWJuUnpMbWR2YjJkc1pTNWpiMjB2Ynk5dllYVjBhREl2WVhWMGFDSXNJQ0FpZEc5clpXNWZkWEpwSWpvZ0ltaDBkSEJ6T2k4dmIyRjFkR2d5TG1kdmIyZHNaV0Z3YVhNdVkyOXRMM1J2YTJWdUlpd2dJQ0poZFhSb1gzQnliM1pwWkdWeVgzZzFNRGxmWTJWeWRGOTFjbXdpT2lBaWFIUjBjSE02THk5M2QzY3VaMjl2WjJ4bFlYQnBjeTVqYjIwdmIyRjFkR2d5TDNZeEwyTmxjblJ6SWl3Z0lDSmpiR2xsYm5SZmVEVXdPVjlqWlhKMFgzVnliQ0k2SUNKb2RIUndjem92TDNkM2R5NW5iMjluYkdWaGNHbHpMbU52YlM5eWIySnZkQzkyTVM5dFpYUmhaR0YwWVM5NE5UQTVMMmRqY2kxd2RXeHNKVFF3Y25WdUxXRnBMWEJ5YjJRdWFXRnRMbWR6WlhKMmFXTmxZV05qYjNWdWRDNWpiMjBpZlE9PSIgfSB9LCAiSHR0cEhlYWRlcnMiOiB7ICJVc2VyLUFnZW50IjogIkRvY2tlci1DbGllbnQvMTkuMDMuNSAoZGFyd2luKSIgfX0=
type: kubernetes.io/dockerconfigjson
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: runai-operator
  namespace: runai
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: runai-operator
  namespace: runai
rules:
  - apiGroups:
      - ""
    resources:
      - pods
      - services
      - services/finalizers
      - endpoints
      - persistentvolumeclaims
      - events
      - configmaps
      - secrets
      - serviceaccounts
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - batch
    resources:
      - jobs
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - apps
    resources:
      - deployments
      - daemonsets
      - replicasets
      - statefulsets
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - ""
    resources:
      - namespaces
    verbs:
      - get
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - create
  - apiGroups:
      - monitoring.coreos.com
    resources:
      - servicemonitors
      - alertmanagers
      - prometheusrules
      - prometheuses
      - podmonitors
    verbs:
      - get
      - create
      - delete
  - apiGroups:
      - apps
    resourceNames:
      - runai
    resources:
      - deployments/finalizers
    verbs:
      - update
  - apiGroups:
      - run.ai
    resources:
      - '*'
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - apiextensions.k8s.io
    resources:
      - customresourcedefinitions
    verbs:
      - '*'
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - roles
      - rolebindings
    verbs:
      - '*'
  - apiGroups:
      - batch
    resources:
      - cronjobs
    verbs:
      - get
      - delete
      - create
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: runai-operator
  namespace: runai
rules:
  - apiGroups:
      - ""
    resources:
      - pods
      - services
      - services/finalizers
      - endpoints
      - persistentvolumeclaims
      - events
      - configmaps
      - secrets
      - serviceaccounts
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - batch
    resources:
      - jobs
      - cronjobs
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - apps
    resources:
      - deployments
      - daemonsets
      - replicasets
      - statefulsets
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - ""
    resources:
      - configmaps
    verbs:
      - get
  - apiGroups:
      - apiextensions.k8s.io
    resources:
      - customresourcedefinitions
    verbs:
      - create
      - get
  - apiGroups:
      - policy
    resources:
      - podsecuritypolicies
    verbs:
      - get
      - delete
      - create
  - apiGroups:
      - storage.k8s.io
    resources:
      - storageclasses
    verbs:
      - get
      - delete
      - create
      - list
      - watch
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - clusterroles
      - clusterrolebindings
    verbs:
      - get
      - delete
      - create
  - apiGroups:
      - ""
    resources:
      - services
    verbs:
      - get
      - delete
      - create
  - apiGroups:
      - ""
    resources:
      - limitranges
      - namespaces
      - nodes
      - persistentvolumes
      - replicationcontrollers
      - resourcequotas
    verbs:
      - list
      - watch
  - apiGroups:
      - admissionregistration.k8s.io
    resources:
      - mutatingwebhookconfigurations
      - validatingwebhookconfigurations
    verbs:
      - '*'
  - apiGroups:
      - scheduling.k8s.io
    resources:
      - priorityclasses
    verbs:
      - get
      - delete
      - create
  - apiGroups:
      - certificates.k8s.io
    resources:
      - certificatesigningrequests
    verbs:
      - get
      - watch
      - list
      - create
      - update
      - patch
      - delete
  - apiGroups:
      - certificates.k8s.io
    resources:
      - certificatesigningrequests/approval
      - certificatesigningrequests/status
    verbs:
      - update
  - apiGroups:
      - extensions
    resources:
      - podsecuritypolicies
    verbs:
      - use
  - apiGroups:
      - extensions
    resources:
      - daemonsets
      - deployments
      - ingresses
      - replicasets
    verbs:
      - list
      - watch
  - apiGroups:
      - networking.k8s.io
    resources:
      - ingresses
    verbs:
      - list
      - watch
  - apiGroups:
      - policy
    resources:
      - poddisruptionbudgets
    verbs:
      - list
      - watch
  - apiGroups:
      - autoscaling
    resources:
      - horizontalpodautoscalers
    verbs:
      - list
      - watch
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: runai-operator
  namespace: runai
subjects:
  - kind: ServiceAccount
    name: runai-operator
roleRef:
  kind: Role
  name: runai-operator
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: runai-operator
  namespace: runai
subjects:
  - kind: ServiceAccount
    name: runai-operator
    namespace: runai
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: departments.scheduling.incubator.k8s.io
  annotations:
    "helm.sh/hook-weight": "1"
spec:
  group: scheduling.incubator.k8s.io
  names:
    kind: Department
    plural: departments
  scope: Cluster
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          properties:
            deservedGpus:
              format: float
              type: number
          type: object
      type: object
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: runaijobs.run.ai
spec:
  group: run.ai
  version: v1
  scope: Namespaced
  names:
    plural:  runaijobs
    singular: runaijob
    kind: RunaiJob
    shortNames:
      - rj
  subresources:
    status: {}
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: projects.run.ai
  annotations:
    "helm.sh/hook-weight": "1"
spec:
  group: run.ai
  version: v1
  scope: Cluster
  names:
    kind: Project
    plural: projects
    singular: project
    shortNames:
      - rp
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          properties:
            department:
              type: string
            deservedGpus:
              format: float
              type: number
            interactiveJobTimeLimitSecs:
              format: int64
              type: integer
            nodeAffinityInteractive:
              items:
                type:
                  string
              type: array
            nodeAffinityTrain:
              items:
                type:
                  string
              type: array
          type: object
      type: object
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: podgroups.scheduling.incubator.k8s.io
  annotations:
    "helm.sh/hook-weight": "2"
spec:
  group: scheduling.incubator.k8s.io
  names:
    kind: PodGroup
    plural: podgroups
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          properties:
            minMember:
              format: int32
              type: integer
            queue:
              type: string
            priorityClassName:
              type: string
          type: object
        status:
          properties:
            succeeded:
              format: int32
              type: integer
            failed:
              format: int32
              type: integer
            running:
              format: int32
              type: integer
          type: object
      type: object
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: queues.scheduling.incubator.k8s.io
  annotations:
    "helm.sh/hook-weight": "1"
spec:
  group: scheduling.incubator.k8s.io
  names:
    kind: Queue
    plural: queues
  scope: Cluster
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          properties:
            deservedGpus:
              format: float
              type: number
            interactiveJobTimeLimitSecs:
              format: int64
              type: integer
            nodeAffinityInteractive:
              items:
                type:
                  string
              type: array
            nodeAffinityTrain:
              items:
                type:
                  string
              type: array
          type: object
      type: object
  version: v1alpha1
---
`
