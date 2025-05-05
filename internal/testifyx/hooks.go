package testifyx

func (ts *TestSuite) BeforeEach(fn func()) *TestSuite {
	ts.beforeEach = fn
	return ts
}

func (ts *TestSuite) AfterEach(fn func()) *TestSuite {
	ts.afterEach = fn
	return ts
}

func (ts *TestSuiteBench) BeforeEach(fn func()) *TestSuiteBench {
	ts.beforeEach = fn
	return ts
}

func (ts *TestSuiteBench) AfterEach(fn func()) *TestSuiteBench {
	ts.afterEach = fn
	return ts
}
