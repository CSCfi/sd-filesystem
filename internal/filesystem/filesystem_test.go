package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sda-filesystem/internal/logs"
	"sort"
	"testing"

	"github.com/billziss-gh/cgofuse/fuse"
)

var testFuse = `{
	"name": "",
	"size": 450,
    "children": [
        {
			"name": "child_1",
			"size": 200,
			"children": [
				{
					"name": "kansio",
					"size": 88,
					"children": [
						{
							"name": "file_1",
							"size": 23,
							"children": null
						},
						{
							"name": "file_2",
							"size": 45,
							"children": null
						},
						{
							"name": "file_3",
							"size": 20,
							"children": null
						}
					] 
				},
				{
					"name": "dir",
					"size": 112,
					"children": [
						{
							"name": "folder",
							"size": 112,
							"children": []
						}
					] 
				}
			] 
		},
		{
			"name": "child_2",
			"size": 250,
			"children": [
				{
					"name": "dir",
					"size": 151,
					"children": []
				},
				{
					"name": "folder",
					"size": 99,
					"children": [
						{
							"name": "file_1",
							"size": 3,
							"children": null
						},
						{
							"name": "file_2",
							"size": 11,
							"children": null
						},
						{
							"name": "test",
							"size": 62,
							"children": null
						},
						{
							"name": "FILE_1_test",
							"size": 23,
							"children": null
						}
					]
				}
			]
		}
	]
}`

type jsonNode struct {
	Name     string      `json:"name"`
	Size     int64       `json:"size"`
	Children *[]jsonNode `json:"children"`
}

func TestMain(m *testing.M) {
	logs.SetSignal(func(i int, s []string) {})
	os.Exit(m.Run())
}

// getMockFuse returns a *Connectfs filled in based on variable testFuse
func getNewFuse(t *testing.T) (fs *Fuse) {
	var nodes jsonNode
	if err := json.Unmarshal([]byte(testFuse), &nodes); err != nil {
		t.Fatal("Could not unmarshal json")
	}

	fs = &Fuse{}
	fs.root = &node{}
	fs.root.stat.Mode = fuse.S_IFDIR | sRDONLY
	fs.root.stat.Size = nodes.Size
	fs.root.chld = map[string]*node{}

	assignChildren(fs.root, nodes.Children)
	return
}

func assignChildren(n *node, template *[]jsonNode) {
	for i, child := range *template {
		n.chld[child.Name] = &node{}
		n.chld[child.Name].stat.Size = child.Size

		if child.Children != nil {
			n.chld[child.Name].stat.Mode = fuse.S_IFDIR | sRDONLY
			n.chld[child.Name].chld = map[string]*node{}
			assignChildren(n.chld[child.Name], (*template)[i].Children)
		} else {
			n.chld[child.Name].stat.Mode = fuse.S_IFREG | sRDONLY
		}
	}
}

func isSameFuse(fs1 map[string]*node, fs2 map[string]*node, path string) error {
	// Names of children of fs1 and fs2
	keys1 := make([]string, len(fs1))
	keys2 := make([]string, len(fs2))

	i := 0
	for k := range fs1 {
		keys1[i] = k
		i++
	}

	i = 0
	for k := range fs2 {
		keys2[i] = k
		i++
	}

	sort.Strings(keys1)
	sort.Strings(keys2)

	if !reflect.DeepEqual(keys1, keys2) {
		return fmt.Errorf("Children differ at node %q. Should be %q, got %q", path, keys1, keys2)
	}

	for _, k := range keys1 {
		if fs2[k].stat.Mode != fs1[k].stat.Mode {
			return fmt.Errorf("Mode not correct at node %q", path+"/"+k)
		}
		if fs2[k].stat.Size != fs1[k].stat.Size {
			return fmt.Errorf("Size not correct at node %q. Expected %d, got %d", path+"/"+k, fs1[k].stat.Size, fs2[k].stat.Size)
		}

		if fuse.S_IFDIR == fs1[k].stat.Mode&fuse.S_IFMT {
			if fs2[k].chld == nil {
				return fmt.Errorf("chld field not initialized at node %q", path+"/"+k)
			}
			if err := isSameFuse(fs1[k].chld, fs2[k].chld, path+"/"+k); err != nil {
				return err
			}
		}
	}

	return nil
}

/*func TestCreateFilesystem(t *testing.T) {
	origFs := getNewFuse(t)

	origGetProjects := api.GetProjects
	origGetContainers := api.GetContainers
	origRemoveInvalidChars := removeInvalidChars
	origCreateObjects := createObjects
	origNewNode := newNode

	defer func() {
		api.GetProjects = origGetProjects
		api.GetContainers = origGetContainers
		removeInvalidChars = origRemoveInvalidChars
		createObjects = origCreateObjects
		newNode = origNewNode
	}()

	api.GetProjects = func() ([]api.Metadata, error) {
		return []api.Metadata{{Bytes: 200, Name: "child_1"}, {Bytes: 250, Name: "child_2"}}, nil
	}
	api.GetContainers = func(project string) ([]api.Metadata, error) {
		if project == "child_1" {
			return []api.Metadata{{Bytes: 88, Name: "kansio"}, {Bytes: 112, Name: "dir"}}, nil
		} else if project == "child_2" {
			return []api.Metadata{{Bytes: 151, Name: "dir"}, {Bytes: 99, Name: "folder"}}, nil
		} else {
			return nil, errors.New("Failed to get containers, incorrect project name")
		}
	}
	removeInvalidChars = func(str string, ignore ...string) string {
		return str
	}
	createObjects = func(id int, jobs <-chan containerInfo, wg *sync.WaitGroup, send ...chan<- LoadProjectInfo) {
		defer wg.Done()

		for j := range jobs {
			project := j.project
			container := j.container

			if _, ok := origFs.root.chld[project]; !ok {
				t.Errorf("Invalid project name %q in containerInfo", project)
				continue
			}
			if _, ok := origFs.root.chld[project].chld[container]; !ok {
				t.Errorf("Invalid container name %q in containerInfo", project+"/"+container)
				continue
			}
			if _, ok := origFs.root.chld[project]; !ok {
				t.Errorf("Project %q does not exist in fuse", project)
				continue
			}
			if _, ok := origFs.root.chld[project].chld[container]; !ok {
				t.Errorf("Container %q does not exist in fuse", project+"/"+container)
				continue
			}

			for key, value := range origFs.root.chld[project].chld[container].chld {
				j.fs.root.chld[project].chld[container].chld[key] = value
			}
		}
	}
	newNode = func(dev uint64, ino uint64, mode uint32, uid uint32, gid uint32, tmsp fuse.Timespec) *node {
		n := &node{}
		n.stat.Mode = mode
		n.chld = map[string]*node{}
		return n
	}

	ret := CreateFileSystem()
	if ret == nil || ret.root == nil {
		t.Fatal("Filesystem or root is nil")
	}
	if ret.root.stat.Mode != fuse.S_IFDIR|sRDONLY {
		t.Error("Root mode is incorrect")
	}
	if ret.root.stat.Size != origFs.root.stat.Size {
		t.Errorf("Root node size incorrect. Expected %d, got %d", origFs.root.stat.Size, ret.root.stat.Size)
	}
	if err := isSameFuse(origFs.root.chld, ret.root.chld, ""); err != nil {
		t.Fatalf("FUSE was not created correctly: %s", err.Error())
	}
}

func TestCreateObjects(t *testing.T) {
	origFs := getNewFuse(t)
	fs := getNewFuse(t)

	origGetObjects := api.GetObjects
	defer func() { api.GetObjects = origGetObjects }()

	project := "child_2"
	container := "dir"

	api.GetObjects = func(p, c string) ([]api.Metadata, error) {
		if p != project {
			t.Fatalf("GetObject received incorrect project %q, expected %q", p, project)
		}
		if c != container {
			t.Fatalf("GetObject received incorrect container %q, expected %q", p+"/"+c, project+"/"+container)
		}
		return []api.Metadata{{Bytes: 30, Name: "dir1/dir2/dir3/file"}, {Bytes: 50, Name: "dir1/dir4/another_file"},
			{Bytes: 101, Name: "dir1/dir2/logs"}, {Bytes: 1, Name: "dir1/dir5/"}}, nil
	}

	var wg sync.WaitGroup
	jobs := make(chan containerInfo, 1)
	wg.Add(1)
	go createObjects(0, jobs, &wg)
	jobs <- containerInfo{project: "child_2", container: "dir", timestamp: fuse.Timespec{}, fs: fs}
	close(jobs)
	wg.Wait()

	// Adding manually all the neccessary nodes to origFuse
	origFs.root.chld[project].chld[container].chld["dir1"] = &node{}
	origFs.root.chld[project].chld[container].chld["dir1"].stat.Mode = fuse.S_IFDIR | sRDONLY
	origFs.root.chld[project].chld[container].chld["dir1"].stat.Size = 181
	origFs.root.chld[project].chld[container].chld["dir1"].chld = map[string]*node{}

	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"] = &node{}
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].stat.Mode = fuse.S_IFDIR | sRDONLY
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].stat.Size = 131
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld = map[string]*node{}

	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir4"] = &node{}
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir4"].stat.Mode = fuse.S_IFDIR | sRDONLY
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir4"].stat.Size = 50
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir4"].chld = map[string]*node{}

	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld["dir3"] = &node{}
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld["dir3"].stat.Mode = fuse.S_IFDIR | sRDONLY
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld["dir3"].stat.Size = 30
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld["dir3"].chld = map[string]*node{}

	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld["logs"] = &node{}
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld["logs"].stat.Mode = fuse.S_IFREG | sRDONLY
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld["logs"].stat.Size = 101

	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld["dir3"].chld["file"] = &node{}
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld["dir3"].chld["file"].stat.Mode = fuse.S_IFREG | sRDONLY
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir2"].chld["dir3"].chld["file"].stat.Size = 30

	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir4"].chld["another_file"] = &node{}
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir4"].chld["another_file"].stat.Mode = fuse.S_IFREG | sRDONLY
	origFs.root.chld[project].chld[container].chld["dir1"].chld["dir4"].chld["another_file"].stat.Size = 50

	if err := isSameFuse(origFs.root.chld, fs.root.chld, ""); err != nil {
		t.Errorf("Objects not added correctly: %s", err.Error())
	}
}

func TestCreateObjects_Error(t *testing.T) {
	origFs := getNewFuse(t)
	fs := getNewFuse(t)

	origGetObjects := api.GetObjects
	defer func() { api.GetObjects = origGetObjects }()

	api.GetObjects = func(string, string) ([]api.Metadata, error) {
		return nil, errors.New("Error occured")
	}

	var wg sync.WaitGroup
	jobs := make(chan containerInfo, 1)
	wg.Add(1)
	go createObjects(0, jobs, &wg)
	jobs <- containerInfo{project: "child_2", container: "dir", timestamp: fuse.Timespec{}, fs: fs}
	close(jobs)
	wg.Wait()

	if err := isSameFuse(origFs.root.chld, fs.root.chld, ""); err != nil {
		t.Errorf("Fuse should not have been modified: %s", err.Error())
	}
}

func TestRemoveInvalidChars(t *testing.T) {
	var tests = []struct {
		original, modified, ignore string
	}{
		{"b.a:d!s/t_r@i+n|g", "b.a.d.s_t_r_i_n_g", ""},
		{"qwerty__\"###hello<html>$$money$$", "qwerty__.___hello.html.__money__", ""},
		{"qwerty__\"###hello<html>$$money$$", "qwerty__.###hello.html.__money__", "#"},
		{"%_csc::>d>p>%%'hello'", "__csc...d.p.__.hello.", ""},
		{"%_csc::>d>p>%%'hello'", "%_csc...d.p.%%.hello.", "%"},
	}

	for i, tt := range tests {
		testname := fmt.Sprintf("REMOVE_%d", i+1)
		t.Run(testname, func(t *testing.T) {
			var ret string
			if tt.ignore != "" {
				ret = removeInvalidChars(tt.original, tt.ignore)
			} else {
				ret = removeInvalidChars(tt.original)
			}

			if ret != tt.modified {
				t.Errorf("String %q should have become %q, got %q", tt.original, tt.modified, ret)
			}
		})
	}
}

func TestLookupNode(t *testing.T) {
	fs := getNewFuse(t)

	var tests = []struct {
		testname, path       string
		prntMatch, nodeMatch *node
		dir, clash           bool
	}{
		// File already exists
		{
			"FILE_EXISTS", "child_2/folder/file_2",
			fs.root.chld["child_2"].chld["folder"], fs.root.chld["child_2"].chld["folder"].chld["file_2"],
			false, false,
		},
		// Folder already exists
		{
			"FOLDER_EXISTS", "child_1/dir/folder",
			fs.root.chld["child_1"].chld["dir"], fs.root.chld["child_1"].chld["dir"].chld["folder"],
			true, false,
		},
		// Parent does not exist
		{
			"INVALID_PATH", "child_1/folder/file_2", nil, nil, false, false,
		},
		// A directory with the same name already exists
		{
			"MATCHING_DIR", "child_1/dir/folder",
			fs.root.chld["child_1"].chld["dir"], fs.root.chld["child_1"].chld["dir"].chld["folder"],
			false, true,
		},
		// A file with the same name already exists
		{
			"MATCHING_FILE", "child_1/kansio/file_3",
			fs.root.chld["child_1"].chld["kansio"], fs.root.chld["child_1"].chld["kansio"].chld["file_3"],
			true, true,
		},
		// OK
		{
			"OK_1", "child_2/folder/file_3",
			fs.root.chld["child_2"].chld["folder"], nil,
			false, false,
		},
		// OK
		{
			"OK_2", "child_1//dir/folder///another_folder",
			fs.root.chld["child_1"].chld["dir"].chld["folder"], nil,
			true, false,
		},
	}

	for _, tt := range tests {
		testname := tt.testname
		t.Run(testname, func(t *testing.T) {
			prnt, name, node, isDir := lookupNode(fs, tt.path)

			if prnt == nil {
				if tt.prntMatch != nil {
					t.Errorf("Parent node was nil for path %q", tt.path)
				}
				return
			}

			if tt.prntMatch != prnt {
				t.Errorf("Parent node incorrect for path %q. Expected address %p, got %p", tt.path, tt.prntMatch, prnt)
				return
			}

			if name != path.Base(tt.path) {
				t.Errorf("Name incorrect for path %q. Expected %q, got %q", tt.path, path.Base(tt.path), name)
				return
			}

			if node == nil {
				if tt.nodeMatch != nil {
					t.Errorf("Node nil for path %q", tt.path)
				}
				return
			}

			if node != tt.nodeMatch {
				t.Errorf("Node incorrect for path %q. Expected address %p, got %p", tt.path, tt.nodeMatch, node)
				return
			}

			if tt.clash {
				if tt.dir && isDir {
					t.Errorf("LookupNode found a matching directory, not a matching file, for path %q.", tt.path)
				} else if !tt.dir && !isDir {
					t.Errorf("LookupNode found a matching file, not a matching directory, for path %q.", tt.path)
				}
			} else {
				if !tt.dir && isDir {
					t.Errorf("LookupNode found a matching directory, not a matching file, for path %q.", tt.path)
				} else if tt.dir && !isDir {
					t.Errorf("LookupNode found a matching file, not a matching directory, for path %q.", tt.path)
				}
			}
		})
	}
}

func TestNewNode(t *testing.T) {
	var tests = []struct {
		dir  bool
		ino  uint64
		tmsp fuse.Timespec
	}{
		{true, 32, fuse.Now()},
		{false, 57, fuse.NewTimespec(time.Now().AddDate(0, 0, -1))},
	}

	for i, tt := range tests {
		testname := fmt.Sprintf("NEW_NODE_%d", i+1)
		t.Run(testname, func(t *testing.T) {
			var mode uint32
			if tt.dir {
				mode = fuse.S_IFDIR | sRDONLY
			} else {
				mode = fuse.S_IFREG | sRDONLY
			}

			node := newNode(0, tt.ino, mode, 0, 0, tt.tmsp)

			if node == nil {
				t.Error("Node is nil")
				return
			}

			if node.stat.Ino != tt.ino {
				t.Errorf("File serial number incorrect. Expected %d, got %d", tt.ino, node.stat.Ino)
			} else if node.stat.Mode != mode {
				t.Errorf("Mode incorrect. Expected %d, got %d", mode, node.stat.Mode)
			} else if node.stat.Atim != tt.tmsp {
				t.Errorf("Atim field incorrect. Expected %q, got %q", tt.tmsp.Time().String(), node.stat.Atim.Time().String())
			} else if node.stat.Ctim != tt.tmsp {
				t.Errorf("Ctim field incorrect. Expected %q, got %q", tt.tmsp.Time().String(), node.stat.Ctim.Time().String())
			} else if node.stat.Mtim != tt.tmsp {
				t.Errorf("Mtim field incorrect. Expected %q, got %q", tt.tmsp.Time().String(), node.stat.Mtim.Time().String())
			} else if node.stat.Birthtim != tt.tmsp {
				t.Errorf("Birthtim field incorrect. Expected %q, got %q", tt.tmsp.Time().String(), node.stat.Birthtim.Time().String())
			} else if tt.dir && node.chld == nil {
				t.Errorf("Node's chld field was not initialized")
			}
		})
	}
}

func TestMakeNode(t *testing.T) {
	var tests = []struct {
		mockLookupNode func(*Connectfs, string) (*node, string, *node, bool)
		expectedOutput int
		size           int64
		dir            bool
		rename         string
		path, testname string
	}{
		// Inserting file to invalid path
		{
			func(fs *Connectfs, path string) (*node, string, *node, bool) {
				return nil, "test", nil, false
			},
			-fuse.ENOENT, 56, false,
			"", "folder/test", "INVALID_PATH_1",
		},
		// Inserting folder to invalid path
		{
			func(fs *Connectfs, path string) (*node, string, *node, bool) {
				return nil, "test2", &node{}, true
			},
			-fuse.ENOENT, 23, false,
			"", "folder/test2", "INVALID_PATH_2",
		},
		// Inserting file where file already exists
		{
			func(fs *Connectfs, path string) (*node, string, *node, bool) {
				prnt := fs.root.chld["child_2"].chld["folder"]
				node := fs.root.chld["child_2"].chld["folder"].chld["file_1"]
				return prnt, "file_1", node, false
			},
			-fuse.EEXIST, 9, false,
			"", "child_2/folder/file_1", "FILE_EXISTS",
		},
		// Inserting folder where folder already exists
		{
			func(fs *Connectfs, path string) (*node, string, *node, bool) {
				prnt := fs.root.chld["child_1"].chld["dir"]
				node := fs.root.chld["child_1"].chld["dir"].chld["folder"]
				return prnt, "folder", node, true
			},
			-fuse.EEXIST, 345, true,
			"", "child_1/dir/folder", "FOLDER_EXISTS",
		},
		// Inserting file where folder with same name already exists
		{
			func(fs *Connectfs, path string) (*node, string, *node, bool) {
				prnt := fs.root.chld["child_2"]
				node := fs.root.chld["child_2"].chld["folder"]
				return prnt, "folder", node, true
			},
			0, 2, false,
			"FILE_1_folder", "child_2/folder", "MATCHING_FOLDER_EXISTS",
		},
		// Inserting folder where file with same name already exists
		{
			func(fs *Connectfs, path string) (*node, string, *node, bool) {
				prnt := fs.root.chld["child_2"].chld["folder"]
				node := fs.root.chld["child_2"].chld["folder"].chld["test"]
				return prnt, "test", node, false
			},
			0, 123, true,
			"FILE_2_test", "child_2/folder/test", "MATCHING_FILE_EXISTS",
		},
		// OK
		{
			func(fs *Connectfs, path string) (*node, string, *node, bool) {
				prnt := fs.root.chld["child_1"].chld["dir"]
				return prnt, "n", nil, true
			},
			0, 89, false,
			"", "child_1/dir/n", "OK_1",
		},
		// OK
		{
			func(fs *Connectfs, path string) (*node, string, *node, bool) {
				prnt := fs.root.chld["child_1"]
				return prnt, "newnode", nil, true
			},
			0, 45, true,
			"", "child_1/newnode", "OK_2",
		},
	}

	origLookupNode := lookupNode
	origNewNode := newNode

	defer func() {
		lookupNode = origLookupNode
		newNode = origNewNode
	}()

	newNode = func(dev uint64, ino uint64, mode uint32, uid uint32, gid uint32, tmsp fuse.Timespec) *node {
		n := &node{}
		n.stat.Mode = mode
		n.chld = map[string]*node{}
		return n
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			var mode uint32
			if tt.dir {
				mode = fuse.S_IFDIR | sRDONLY
			} else {
				mode = fuse.S_IFREG | sRDONLY
			}

			origFS := getNewFuse(t)
			fs := getNewFuse(t)

			lookupNode = tt.mockLookupNode
			ret := fs.makeNode(tt.path, mode, 0, tt.size, fuse.Timespec{})

			if tt.expectedOutput != ret {
				t.Errorf("Incorrect return value, expected %d, got %d", tt.expectedOutput, ret)
				return
			}

			renamed := map[string]string{}

			if ret == 0 {
				nodes := strings.Split(tt.path, "/")
				follow := origFS.root
				for i := range nodes[:len(nodes)-1] {
					follow = follow.chld[nodes[i]]
				}

				if tt.rename == "" {
					follow.chld[nodes[len(nodes)-1]] = &node{}
					follow.chld[nodes[len(nodes)-1]].stat.Mode = mode
					follow.chld[nodes[len(nodes)-1]].stat.Size = tt.size

					if tt.dir {
						follow.chld[nodes[len(nodes)-1]].chld = map[string]*node{}
					}
				} else {
					follow.chld[tt.rename] = &node{}
					follow.chld[tt.rename].stat.Mode = fuse.S_IFREG | sRDONLY
					follow.chld[nodes[len(nodes)-1]].stat.Mode = fuse.S_IFDIR | sRDONLY

					if !tt.dir {
						follow.chld[tt.rename].stat.Size = tt.size
					} else {
						follow.chld[tt.rename].stat.Size = follow.chld[nodes[len(nodes)-1]].stat.Size
						follow.chld[nodes[len(nodes)-1]].stat.Size = tt.size
					}

					nodes[len(nodes)-1] = tt.rename
					renamed[strings.Join(nodes, "/")] = tt.path
				}
			}

			if err := isSameFuse(origFS.root.chld, fs.root.chld, ""); err != nil {
				t.Errorf("FUSE did not change correctly: %s", err.Error())
			}

			if !reflect.DeepEqual(renamed, fs.renamed) {
				t.Errorf("List of renamed nodes incorrect. Expected %q, got %q", renamed, fs.renamed)
			}
		})
	}
}
*/
