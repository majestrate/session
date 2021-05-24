package swarm


var seedNodes = []ServiceNode{
	ServiceNode{
		RemoteIP: "public.loki.foundation",
		StoragePort: 443,
	},
}

func WithSeedNodes(visit func(ServiceNode)) {
	for _, node := range seedNodes {
		visit(node)
	}
}
