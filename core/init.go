package core

func Init(fn func() []ItemHolder) {
	for _, holder := range fn() {
		holder.Startup()
	}
}
