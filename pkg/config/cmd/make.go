package main

<<<<<<< HEAD
<<<<<<< HEAD
import "github.com/QinYuuuu/abvss/pkg/config"
=======
import "abvss/pkg/config"
>>>>>>> 19b0d27 (Initial commit)
=======
import "abvss/pkg/config"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b

func main() {
	//config.ElgamalCurve25519KeyGen(128, "D:/mycode/gocode/src/abvss")
	//config.SigKeyGen(8, 3, "D:/mycode/gocode/src/abvss")
	//config.EncKeyGen(8, 3, "D:/mycode/gocode/src/abvss")
	//config.LoadSigKey(4, "D:/mycode/gocode/src/abvss")
	//ip1, ip2, ip3 := config.LoadIPList_Local(4, "D:/mycode/gocode/src/abvss")
	//fmt.Println(ip1, ip2, ip3)
	//config.GenerateIPList_Local(4, "D:/mycode/gocode/src/abvss")
	//config.LoadIPList_Local(4, "D:/mycode/gocode/src/abvss")
	config.LoadRawIP_2(8, 4, "D:/mycode/gocode/src/abvss")
	config.SHRawIP(8)
}
