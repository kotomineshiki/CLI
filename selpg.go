package main


import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	flag "github.com/spf13/pflag"
)

type Args struct {
	s         int
	e         int
	l         int
	f         bool//按照分页符\f分页的标识符
	d         string//传输目的地
	inputFile string//打开文件
}
func getArgs(args *Args) {//读取
	flag.IntVarP(&args.s, "start", "s", 0, "start")//-s
	flag.IntVarP(&args.e, "end", "e", 0, "end")//-e
	flag.IntVarP(&args.l, "line", "l", -1, "line")//
	flag.BoolVarP(&args.f, "final", "f", false, "final")
	flag.StringVarP(&args.d, "destination", "d", "", "destination")
	flag.Parse()
	otherArgs := flag.Args()
	if len(otherArgs) > 0 {
		args.inputFile = otherArgs[0]
	} else {
		args.inputFile = ""
	}
	checkArgs(args)//检验合法性
}
func checkArgs(args *Args) {//合法性检验
	if args.s == 0 || args.e == 0 {
		os.Stderr.Write([]byte("Please input -s and -e\n"))
		os.Exit(0)
	}
	if args.s > args.e {
		os.Stderr.Write([]byte("Invalid input about -s and -e\n"))
		os.Exit(0)
	}
	if args.f && args.l != -1 {
		os.Stderr.Write([]byte("Please choose either -f or -l\n"))
		os.Exit(0)
	}
}

func getReader(args *Args) *bufio.Reader {
	var reader *bufio.Reader
	if args.inputFile == "" {
		reader = bufio.NewReader(os.Stdin)
	} else {
		file, err := os.Open("./" + args.inputFile)
		if err != nil {
			os.Stderr.Write([]byte("File does not exist\n"))
			os.Exit(1)
		}
		reader = bufio.NewReader(file)
	}
	return reader
}
func executeArgs(args *Args) {
	var reader *bufio.Reader
	reader = getReader(args)//获取reader
	//get writer
	if args.d == "" {
		writer := bufio.NewWriter(os.Stdout)
		if args.f {
			readByF(args, reader, writer)
		} else {
			if args.l == -1 {
				args.l = 72//默认七十二行
			}
			readByLine(args, reader, writer)//selpg -s1 -e1 -l
		}
	} else {
		var cmd = exec.Command("./" + args.d)
		writer, err := cmd.StdinPipe()//通过管道连接子进程
		if err != nil {
			fmt.Println("Error", err)
			os.Exit(1)
		}
		if err := cmd.Start(); err != nil {
			fmt.Println("Error", err)
			os.Exit(1)
		}
		if args.f {
			readByFWithDestination(args, reader, writer)//按照分页符读取selpg -s1 -e1 -f
		} else {
			if args.l == -1 {
				args.l = 72
			}
			readByLWithDestination(args, reader, writer)//selpg -s1 -e1 -l [process_name]
		}
		writer.Close()
		if err := cmd.Wait(); err != nil {
			fmt.Println("Error")
			os.Exit(1)
		}
	}
}
func readByLine(args *Args, reader *bufio.Reader, writer *bufio.Writer) {//从start读到end
	for  i := 1; i <= args.e; i++ {
		if i < args.s {
			for lineCount := 0; lineCount < args.l; lineCount++ {
				reader.ReadBytes('\n')
			}
		} else {
			for lineCount := 0; lineCount < args.l; lineCount++ {
				line, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						writer.WriteByte('\n')
						writer.Flush()
						break
					}
					os.Stderr.Write([]byte("Read failed\n"))
					os.Exit(1)
				}
				writer.Write(line)
				writer.Flush()
			}
		}
	}
}

func readByF(args *Args, reader *bufio.Reader, writer *bufio.Writer) {
	for pageCount := 1; pageCount <= args.e; pageCount++ {
		for {
			char, err := reader.ReadByte()
			if char == '\f' {
				break
			}
			if err != nil {
				if err == io.EOF {
					writer.WriteByte('\n')
					writer.Flush()
					break
				}
				os.Stderr.Write([]byte("Read failed\n"))
				os.Exit(1)
			}
			if pageCount >= args.s {
				writer.WriteByte(char)
				writer.Flush()//清空buffer
			}
		}
	}
}

func readByLWithDestination(args *Args, reader *bufio.Reader, writer io.WriteCloser) {
	for pageCount := 1; pageCount <= args.e; pageCount++ {
		if pageCount < args.s {
			for lineCount := 0; lineCount < args.l; lineCount++ {
				reader.ReadBytes('\n')
			}
		} else {
			for lineCount := 0; lineCount < args.l; lineCount++ {
				line, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					os.Stderr.Write([]byte("Read failed\n"))
					os.Exit(1)
				}
				writer.Write(line)
			}
		}
	}
}

func readByFWithDestination(args *Args, reader *bufio.Reader, writer io.WriteCloser) {
	for pageCount := 1; pageCount <= args.e; pageCount++ {
		for {
			char, err := reader.ReadByte()
			if char == '\f' {
				break
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				os.Stderr.Write([]byte("Read failed\n"))
				os.Exit(1)
			}
			if pageCount >= args.s {
				writer.Write([]byte{char})
			}
		}
	}
}
func main() {
	args := new(Args)
	getArgs(args)//读取字符串
	executeArgs(args)//执行
}