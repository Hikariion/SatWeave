package task

// worker 中实际执行的 SubTask
/*
   rpc   == data ==>  InputReceiver
         == data ==>  channel
         == data ==>  Core
         == data ==>  channel
         == data ==>  OutputDispenser
         == data ==>  rpc
*/
