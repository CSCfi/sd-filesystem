package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"sda-filesystem/internal/airlock"
	"sda-filesystem/internal/api"
	"sda-filesystem/internal/logs"

	"golang.org/x/term"
)

type keyFlags []string

func (k *keyFlags) String() string {
	return fmt.Sprintf("%v", *k)
}

func (k *keyFlags) Set(value string) error {
	*k = append(*k, value)

	return nil
}

func usage(selfPath string) {
	fmt.Println("Password is read from environment variable CSC_PASSWORD")
	fmt.Println("If this variable is empty airlock requests the password interactively")
	fmt.Println("Usage:")
	fmt.Println(" ", selfPath, "[-segment-size=sizeInMb] "+
		"[-journal-number=journalNumber] [-original-file=unecryptedFilename] "+
		"[-public-key=publicKeyFile] [-quiet] username container filename")
	fmt.Println("Examples:")
	fmt.Println(" ", selfPath, "testuser testcontainer path/to/file")
	fmt.Println(" ", selfPath, "-segment-size=100 -public-key=encryption-key.pub "+
		"testuser testcontainer path/to/file")
	fmt.Println(" ", selfPath, "-segment-size=100 "+
		"-original-file=/path/to/original/unecrypted/file -journal-number=example124 "+
		"-public-key=encryption-key.pub -public-key=another-key.pub testuser testcontainer path/to/file")
}

func main() {
	var publicKeys keyFlags
	flag.Var(&publicKeys, "public-key", "Public key used to encrypt file. Can be given multiple times.")

	segmentSizeMb := flag.Int("segment-size", 4000,
		"Maximum size of segments in Mb used to upload data. Valid range is 10-4000.")
	journalNumber := flag.String("journal-number", "",
		"Journal Number/Name specific for Findata uploads")
	originalFilename := flag.String("original-file", "",
		"Filename of original unecrypted file when uploading pre-encrypted file from Findata VM")
	project := flag.String("project", "", "SD Connect project if it differs from that in the VM")
	quiet := flag.Bool("quiet", false, "Print only errors")
	debug := flag.Bool("debug", false, "Enable debug prints")

	flag.Parse()

	if flag.NArg() != 3 {
		usage(os.Args[0])
		os.Exit(2)
	}

	username := flag.Arg(0)
	container := flag.Arg(1)
	filename := flag.Arg(2)

	if *segmentSizeMb < 10 || *segmentSizeMb > 4000 {
		logs.Fatal("Valid values for segment size are 10-4000")
	}

	if *debug {
		logs.SetLevel("debug")
	} else if *quiet {
		logs.SetLevel("error")
	}

	err := api.GetCommonEnvs()
	if err != nil {
		logs.Fatal(err)
	}

	err = api.InitializeClient()
	if err != nil {
		logs.Fatal(err)
	}

	if isManager, err := airlock.IsProjectManager(*project); err != nil {
		logs.Fatalf("Unable to determine project manager status: %s", err.Error())
	} else if !isManager {
		logs.Fatal("You cannot use Airlock as you are not the project manager")
	}

	if err = airlock.GetProxy(); err != nil {
		logs.Fatal(err)
	}

	encrypted, err := airlock.CheckEncryption(filename)
	if err != nil {
		logs.Fatalf("Failed to check if file is encrypted: %s", err.Error())
	}

	if !encrypted {
		err = airlock.GetPublicKey(publicKeys)
		if err != nil {
			logs.Fatal(err)
		}
	} else if encrypted && len(publicKeys) > 0 {
		logs.Warningf("Public key arguments ignored, file already encrypted")
	}

	password, passwordFromEnv := os.LookupEnv("CSC_PASSWORD")

	if passwordFromEnv {
		logs.Info("Using password from environment variable CSC_PASSWORD")
	} else {
		fmt.Println("Enter Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			logs.Fatalf("Could not read password: %s", err.Error())
		}
		password = string(bytePassword)
	}

	_ = api.BasicToken(username, password)

	err = airlock.Upload(filename, container, uint64(*segmentSizeMb), *journalNumber, *originalFilename, encrypted)
	if err != nil {
		logs.Fatal(err)
	}
}
