package main

// 必要ライブラリインポート
import (
	"bufio"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

type Data struct {
	Excludes []string `yaml:"excludes"`
	RootDir  string   `yaml:"root"`
	Modules  []Module `yaml:"modules"`
}

type Module struct {
	Path           string   `yaml:"path"`
	Url            string   `yaml:"url"`
	CheckoutTarget string   `yaml:"target"`
	Excludes       []string `yaml:"excludes"`
	IsLock         bool     `yaml:"lock"`
}

type DependData struct {
	Modules []Module `yaml:"modules"`
}

const (
	yml_file        = "./GitModuleFile.yml"
	clone_dir       = ".gmm"
	version         = "0.0.2"
	dependency_file = "GMMDepend.yml"
	github_url      = "https://github.com/hatch7023/GitModuleManager"
)

var cloneRepositories map[string]string
var dependRepositories map[string]Module

func main() {
	if len(os.Args) < 2 {
		help()
		return
	}

	switch os.Args[1] {
	case "init":
		initialize()
	case "initDepend":
		initializeDepend()
	case "sync":
		sync()
	case "resync":
		remove_root()
		sync()
	case "clean":
		clean()
	case "cleanCache":
		cleanCache()
	case "help":
		help()
	case "version":
		fmt.Println("Git Module Manager Version " + version)
	default:
		help()
	}
}

func help() {
	fmt.Println("These are common gmm (Git Module Manager) commands used in various situations:")
	fmt.Println("")
	fmt.Print("\x1b[33m")
	fmt.Println("Available commands")
	fmt.Print("\x1b[0m")
	fmt.Println("  init\t\tcreate an template gmm.yml")
	fmt.Println("  initDepend\t\tcreate an template GMMDepend.yml")
	fmt.Println("  sync\t\tsync git repository")
	fmt.Println("  resync\tdelete root directory and sync git repository")
	fmt.Println("  clean\t\tremove tml define root directory")
	fmt.Println("  cleanCache\tremove $HOME/.ggm directory")
	fmt.Println("  help\t\topen this")
	fmt.Println("")
	fmt.Print("\x1b[33m")
	fmt.Println("Information")
	fmt.Print("\x1b[0m")
	fmt.Println("  " + github_url)
}

func initialize() {
	if is_exist_dir_file(yml_file) {
		fmt.Println("yml file exist.")
		return
	}
	content := []byte(
		`
# rsync exclude
excludes: 
  - LICENSE,
  - README*,

root: .

modules:
  -
    path: Temp # sync path
    url: git@github.com:
    target: master # checkout target branch or tag or hash
    # excludes: 
    #  - 
`)
	ioutil.WriteFile(yml_file, content, os.ModePerm)
	fmt.Println("Created yml file.")
}

func initializeDepend() {
	if is_exist_dir_file(dependency_file) {
		fmt.Println("yml file exist.")
		return
	}
	content := []byte(
		`
modules:
  -
    path: Temp # sync path
    url: git@github.com:
    target: master # checkout target branch or tag or hash
    # excludes: 
    #  - 
`)
	ioutil.WriteFile(dependency_file, content, os.ModePerm)
	fmt.Println("Created yml file.")
}

func clean() {
	data := load_yml()

	fmt.Println("May I Delete " + data.RootDir + ". [Y/n]")

	stdin := bufio.NewScanner(os.Stdin)
	stdin.Scan()
	text := stdin.Text()
	if text == "Y" {
		remove_root()
	}
}

func remove_root() {
	data := load_yml()

	err := os.RemoveAll(data.RootDir)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Deleted " + data.RootDir)
}

func cleanCache() {
	usr, _ := user.Current()
	clone_path := usr.HomeDir + "/" + clone_dir

	fmt.Println("May I Delete " + clone_path + ". [Y/n]")

	stdin := bufio.NewScanner(os.Stdin)
	stdin.Scan()
	text := stdin.Text()
	if text == "Y" {
		err := os.RemoveAll(clone_path)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Deleted " + clone_path)
	}
}

func sync() {

	data := load_yml()

	usr, _ := user.Current()

	// clone先を作成
	clone_path := usr.HomeDir + "/" + clone_dir
	create_dir_not_exist(clone_path)

	cloneRepositories = map[string]string{}
	dependRepositories = map[string]Module{}

	for _, m := range data.Modules {
		if !switch_git_repo(data, m, clone_path) {
			fmt.Println("")
			fmt.Print("\x1b[31m")
			fmt.Println("Skip This repository Sync.")
			fmt.Print("\x1b[0m")
		}
	}

	// 依存リポジトリをDL
	if len(dependRepositories) > 0 {
		sync_depend(data, clone_path)
	}
}

func sync_depend(data Data, clone_path string) {
	fmt.Println("")
	fmt.Print("\x1b[33m")
	fmt.Println("***********************************")
	fmt.Println("*** Sync Dependensy Repository. ***")
	fmt.Println("***********************************")
	fmt.Print("\x1b[0m")
	for k, v := range dependRepositories {
		// Cloneしてた場合はスキップ
		// TODO 依存リポジトリのディレクトリ確認
		_, is_exit := cloneRepositories[k]
		if !is_exit {
			if !switch_git_repo(data, v, clone_path) {
				fmt.Print("\x1b[31m")
				fmt.Println("")
				fmt.Println("Skip This repository Sync.")
				fmt.Print("\x1b[0m")
			}
		} else {
			fmt.Print("\x1b[31m")
			fmt.Println("")
			fmt.Println("Skip Sync. " + v.Url)
			fmt.Print("\x1b[0m")
		}
	}
}

// gitのリポジトリを切り替え
func switch_git_repo(data Data, module Module, path string) bool {
	fmt.Print("\x1b[33m")
	fmt.Println("")
	fmt.Println(module.Url)
	fmt.Println("")
	fmt.Print("\x1b[0m")

	// 重複はしないはず
	cloneRepositories[module.Url] = ""

	if module.IsLock {
		fmt.Println("Module is lock")
		return false
	}

	// URLを分解
	paths := strings.Split(module.Url, "/")
	clone_path := path + "/" + strings.Replace(paths[len(paths)-1], ".git", "", 1)
	// 無かったらCloneする
	if !is_exist_dir_file(clone_path) {
		fmt.Println("Clone:" + clone_path)
		// Cloneできなかったらリターン
		if !git_clone(module.Url, clone_path) {
			return false
		}
	}

	// 今のディレクトリ保存
	prevDir, _ := filepath.Abs(".")
	// 移動
	os.Chdir(clone_path)

	// 依存ファイル確認
	if is_exist_dir_file(dependency_file) {
		save_depend_repository()
	}

	// フェッチ
	git_fetch()

	target := module.CheckoutTarget
	if is_exist_branch(module.CheckoutTarget) {
		target = "origin/" + target
	} else if is_exist_tag(module.CheckoutTarget) {
		target = "refs/tags/" + target
	}
	// 切り替え
	if !git_checkout(target) {
		fmt.Println("Can`t not change target.")
		return false
	}

	fmt.Println("Checkout " + target)
	fmt.Println("")

	// もとに戻す
	os.Chdir(prevDir)

	// Sysnc begin
	path_to := data.RootDir + "/" + module.Path
	if !rsync_files(data.Excludes, module.Excludes, clone_path+"/", path_to) {
		fmt.Println("Sync Error.")
	}

	return true
}

// ディレクトリがあるか確認する
func is_exist_dir_file(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// ディレクトリが無ければ作成する
func create_dir_not_exist(path string) {
	if is_exist_dir_file(path) {
		return
	}
	fmt.Println("Create Directory:" + path)
	err := os.MkdirAll(path, 0777)
	if err != nil {
		panic(err)
	}
}

// exec git clone
func git_clone(repo, path string) bool {
	cmd := exec.Command("git", "clone", "--quiet", repo, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err == nil
}

func git_fetch() {
	cmd := exec.Command("git", "fetch", "--prune")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

// 指定のところへ
func git_checkout(branch string) bool {
	cmd := exec.Command("git", "checkout", branch)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err == nil
}

// ブランチが存在するか確認
func is_exist_branch(branch string) bool {
	cmd := exec.Command("sh", "-c", "git branch -r | grep -v HEAD")
	out, _ := cmd.Output()
	branchs := strings.Split(string(out), "\n")
	for _, b := range branchs {
		if strings.TrimSpace(b) == "origin/"+branch {
			return true
		}
	}
	return false
}

func is_exist_tag(tag string) bool {
	cmd := exec.Command("git", "tag", "-l", tag)
	out, _ := cmd.Output()
	return len(string(out)) > 0
}

func load_yml() Data {
	file, err := ioutil.ReadFile(yml_file)
	if err != nil {
		panic(err)
	}

	var data Data
	err = yaml.Unmarshal(file, &data)
	if err != nil {
		panic(err)
	}
	return data
}

func rsync_files(common_excludes []string, excludes []string, from string, to string) bool {
	create_dir_not_exist(to)

	var args []string
	args = append(args, "-rvu", "--delete-excluded", "--exclude=.git*", "--exclude=.DS_Store", "--exclude=GMMDepend.yml")

	for _, e := range common_excludes {
		args = append(args, "--exclude="+e)
	}
	// 個別に無視が設定されていれば
	if len(excludes) > 0 {
		for _, e := range excludes {
			args = append(args, "--exclude="+e)
		}
	}

	args = append(args, from, to)
	cmd := exec.Command("rsync", args...)

	fmt.Println(strings.Join(cmd.Args, " "))
	fmt.Println("")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err == nil
}

// 依存関係を確保
func save_depend_repository() {
	// エラーのハンドルはナシ
	file, err := ioutil.ReadFile(dependency_file)
	if err != nil {
		panic(err)
		return
	}

	var data DependData
	err = yaml.Unmarshal(file, &data)
	if err != nil {
		return
	}

	if len(data.Modules) <= 0 {
		return
	}

	for _, m := range data.Modules {
		_, is_exit := dependRepositories[m.Url]
		if !is_exit {
			dependRepositories[m.Url] = m
		}
	}
}
