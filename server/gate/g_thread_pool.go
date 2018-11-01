package main

type ThreadPool struct {
	size_ int
	pool_ []*Thread
	index int
}

func (this *ThreadPool) NextIndex() (ret uint32) {
	ret = uint32(this.index % this.size_)
	this.index++
	return ret
}

func (this *ThreadPool) Append(index uint32, id uint32, msg []byte, msg_id uint32) {
	this.pool_[index].Append(id, msg, msg_id)
}

func NewThreadPool(size int) *ThreadPool {
	if size <= 0 {
		panic(size)
	}

	var ret = &ThreadPool{}
	ret.size_ = size
	ret.index = 0
	ret.pool_ = make([]*Thread, size)
	for i := range ret.pool_ {
		ret.pool_[i] = NewThread()
		go ret.pool_[i].Run()
	}

	return ret
}
