package worker

type OutputDispenser struct {
	/*
	                                              / PartitionDispenser
	   SubTask   ->   channel   ->   Dispenser    - PartitionDispenser   ->  SubTask
	                              (SubTaskClient) \ PartitionDispenser
	*/
}
