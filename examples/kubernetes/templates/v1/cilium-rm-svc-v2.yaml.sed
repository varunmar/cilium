---
apiVersion: __DS_API_VERSION__
kind: DaemonSet
metadata:
  name: cilium-rm-svc-v2
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: cilium-rm-svc-v2
      kubernetes.io/cluster-service: "true"
  template:
    metadata:
      labels:
        k8s-app: cilium-rm-svc-v2
        kubernetes.io/cluster-service: "true"
    spec:
      affinity:
        podAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: "k8s-app"
                operator: In
                values:
                - cilium
            topologyKey: "kubernetes.io/hostname"
      containers:
        - name: cilium-rm-svc-v2
          image: docker.io/cilium/cilium:__CILIUM_VERSION__
          imagePullPolicy: IfNotPresent
          command: ["/bin/bash"]
          args:
          - -c
          - "rm /sys/fs/bpf/tc/globals/cilium_lb{4,6}_{services_v2,backends,rr_seq_v2}; touch /tmp/ready; sleep 1h"
          livenessProbe:
            exec:
              command:
              - cat
              - /tmp/ready
            initialDelaySeconds: 5
            periodSeconds: 5
          readinessProbe:
            exec:
              command:
              - cat
              - /tmp/ready
            initialDelaySeconds: 5
            periodSeconds: 5
          volumeMounts:
          - mountPath: /sys/fs/bpf
            name: bpf-maps
      restartPolicy: Always
      tolerations:
        - effect: NoSchedule
          key: node.kubernetes.io/not-ready
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
        - effect: NoSchedule
          key: node.cloudprovider.kubernetes.io/uninitialized
          value: "true"
        - key: CriticalAddonsOnly
          operator: "Exists"
      volumes:
      - hostPath:
          path: /sys/fs/bpf
          type: DirectoryOrCreate
        name: bpf-maps
