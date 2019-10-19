package klog

func ExampleKvLogger() {
	logger := ToKvLogger(Std.WithWriter(DiscardWriter()))
	logger.Infof("test kv logger", "key1", "value1", "key2", "value2", "key3", 123)
	// t=2019-10-19T16:16:31.336054+08:00 lvl=INFO key1=value1 key2=value2 key3=123 msg="test kv logger"

	// Output:
	//
}
