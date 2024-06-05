package main

import "github.com/QinYuuuu/abvss/pkg/config"

func main() {
	//config.ElgamalCurve25519KeyGen(128, "D:/mycode/gocode/src/abvss")
	//config.SigKeyGen(4, 1, "D:/mycode/gocode/src/abvss")
	//config.LoadSigKey(4, "D:/mycode/gocode/src/abvss")
	config.GenerateIPList(4, "D:/mycode/gocode/src/abvss")
}
