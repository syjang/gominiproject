package main

func main() {

	FileEncrypter("t.prototxt", "fe.p")

	FileDecrypter("fe.p", "fe.proto", 104691)
}
