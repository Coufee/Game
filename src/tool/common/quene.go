package common

type Quene struct {
	elems     []interface{}
	nelems    int
	popIndex  int
	pushIndex int
}

func (this *Quene) Len() int {
	return this.nelems
}

func (this *Quene) Push(elem interface{}) {
	if this.nelems == len(this.elems) {
		this.expand()
	}
	this.elems[this.pushIndex] = elem
	this.nelems++
	this.pushIndex = (this.pushIndex + 1) % len(this.elems)
}

func (this *Quene) Pop() (elem interface{}) {
	if this.nelems == 0 {
		return 0
	}
	elem = this.elems[this.popIndex]
	this.elems[this.popIndex] = nil
	this.nelems--
	this.popIndex = (this.popIndex + 1) % len(this.elems)
	return elem
}

func (this *Quene) expand() {
	curcap := len(this.elems)
	var newcap int
	if curcap == 0 {
		newcap = 8
	} else if curcap < 1024 {
		newcap = curcap * 2
	} else {
		newcap = curcap + (curcap / 4)
	}

	elems := make([]interface{}, newcap)

	if this.popIndex == 0 {
		copy(elems, this.elems)
		this.pushIndex = curcap
	} else {
		newpopIndex := newcap - (curcap - this.popIndex)
		copy(elems, this.elems[:this.popIndex])
		copy(elems[newpopIndex:], this.elems[this.popIndex:])
		this.popIndex = newpopIndex
	}

	for i := range this.elems {
		this.elems[i] = nil
	}
	this.elems = elems
}
