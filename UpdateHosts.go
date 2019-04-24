package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type name map[string]string
type instance map[string]name

var HOSTS map[string]string
var n=name{}
var inst=instance{}
var tags=Quote([]string{"key_name","id","tags.Name","private_ip"})
var keys=Escape(tags)
var ID string=""
var IP string=""
var NAME string=""
const ETCHOST string = "/etc/hosts"

func main() {
	r,e:=RunCmd("grep",[]string{"-i",keys,"terraform.tfstate"})
	//check(e,"Grep failed")
	if e==nil {
		scanData(r)
	}
}

func scanData(r string) {
	HOSTS=map[string]string{}
	for _,s:=range strings.Split(r,"\n") {
		t:=strings.Trim(s," ")
		if strings.HasPrefix(t,string(tags[0])) {
			_,NAME=kv(t)
		} else if strings.HasPrefix(t,string(tags[1])) {
			_,ID=kv(t)
			inst[ID]=n
			NAME=""
			IP=""
		} else if strings.HasPrefix(t,string(tags[2])) {
			if NAME=="" {
				_,NAME=kv(t)
			}
		} else if strings.HasPrefix(t,string(tags[3])) {
			_, IP = kv(t)
			n[NAME] = IP
			_, T := HOSTS[NAME]
			if !T {
				HOSTS[NAME] = ""
			}
			HOSTS[NAME] += IP
			HOSTS[NAME] += " "
		} else if t=="" {
		} else {
			log.Println("Unmatched string returned from terraform.tfstate:",t)
		}
	}
	for k,v:=range HOSTS {
		updateHosts(k,v)
	}
}

func updateHosts(n string,ips string) {
	h,e:=GetFile(ETCHOST)
	check(e,"Failed to read /etc/hosts")
	if n=="" || ips=="" {return}
	ips=strings.Trim(ips," ")
	head:=fmt.Sprintf("###DO NOT EDIT THIS LINE vvvvvv### %s ###",n)
	foot:=fmt.Sprintf("###DO NOT EDIT THIS LINE ^^^^^^### %s ###",n)
	i:=-1
	j:=-1
	for k,s:=range h {
		if strings.HasPrefix(s,head) {
			i=k
		}
		if strings.HasPrefix(s,foot) {
			j=k
			if len(h)>k && strings.Trim(h[k+1]," ")=="" {
				j++
			}
		}
	}
	if i!=-1 && j!=-1 {
		h=append(h[:i],h[j+1:]...)
	}
	o,oe:=os.Create(ETCHOST)
	check(oe,"Could not open /etc/hosts for output")
	defer o.Close()
	for _,l:=range h {
		o.WriteString(fmt.Sprintf("%s\n",l))
	}
	o.WriteString(fmt.Sprintf("%s\n",head))
	for i,s:=range strings.Split(ips," ") {
		o.WriteString(fmt.Sprintf("%s\t%s\t%s-%d\n",s,n,n,i+1))
	}
	o.WriteString(fmt.Sprintf("%s\n",foot))
}

func kv(s string) (string,string) {
	as:=strings.Split(s,":")
	k:=strings.Trim(as[0],"\"")
	v:=strings.Trim(as[1]," ")
	v=strings.Trim(v,",")
	v=strings.Trim(v,"\"")
	return k,v
}

func Quote(sa []string) []string {
	s:=[]string{}
	for _,t:=range sa {
		u:=strings.Join([]string{"\"",t,"\""},"")
		s=append(s,u)
	}
	return s
}

func Escape(sa []string) string {
	return strings.Join(sa,"\\|")
}

func GetFile(fn string) ([]string,error) {
	temp,err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(temp),"\n"), err
}

func check(e error,m string) {
	if e!=nil {
		log.Fatal(m,"\n",e)
	}
}

func RunCmd(in string, args []string) (string,error) {
	cmd := exec.Command(in, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	//cmd.Stdout = os.Stdout
	var e bytes.Buffer
	cmd.Stderr = &e
	//cmd.Stderr = os.Stderr
	//cmd.Stdin = os.Stdin
	// SecurityTokenScript -u ndb338  -p xxxxxx
	//out, err := cmd.CombinedOutput()
	err := cmd.Run()
	return out.String(),err
}
