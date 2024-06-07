package main

import "github.com/QinYuuuu/abvss/pkg/config"

func main() {
	//config.ElgamalCurve25519KeyGen(128, "D:/mycode/gocode/src/abvss")
	//config.SigKeyGen(32, 10, "D:/mycode/gocode/src/abvss")
	//config.EncKeyGen(32, 10, "D:/mycode/gocode/src/abvss")
	//config.LoadSigKey(4, "D:/mycode/gocode/src/abvss")
	//ip1, ip2, ip3 := config.LoadIPList_Local(4, "D:/mycode/gocode/src/abvss")
	//fmt.Println(ip1, ip2, ip3)
	//config.GenerateIPList_Local(4, "D:/mycode/gocode/src/abvss")
	//config.LoadIPList_Local(4, "D:/mycode/gocode/src/abvss")
	config.LoadRawIP_2(8, 4, "D:/mycode/gocode/src/abvss")
	config.SHRawIP(8)
}
