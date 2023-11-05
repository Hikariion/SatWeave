package worker

type InputReceiver struct {
	/*
	   PartitionDispenser   \
	   PartitionDispenser   -    subtask   ->   InputReceiver   ->   channel
	   PartitionDispenser   /            (遇到 event 阻塞，类似 Gate)
	*/
	inputChannel chan interface{}
	partitions   []*InputPartitionReceiver
}

type InputPartitionReceiver struct {
	queue chan interface{}
}
