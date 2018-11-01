package main

type xmlConfig struct {
	Jemalloc uint32 `xml:"jemalloc"`
	Log      xmlLog `xml:"log"`
	In       xmlIn  `xml:"in"`
	Out      xmlOut `xml:"out"`
}

type xmlLog struct {
	File  string `xml:"file"`
	Level string `xml:"level"`
}

type xmlIn struct {
	Ip                 string `xml:"ip"`
	Port               uint32 `xml:"port"`
	Thread             uint32 `xml:"thread"`
	Max_read_msg_size  uint32 `xml:"max_read_msg_size"`
	Max_write_msg_size uint32 `xml:"max_write_msg_size"`
}

type xmlOut struct {
	Ip                 string `xml:"ip"`
	Port               uint32 `xml:"port"`
	Thread             uint32 `xml:"thread"`
	Broadcast          uint32 `xml:"broadcast"`
	Max_read_msg_size  uint32 `xml:"max_read_msg_size"`
	Max_write_msg_size uint32 `xml:"max_write_msg_size"`
}
