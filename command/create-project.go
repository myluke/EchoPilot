package command

import (
	"archive/zip"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ei18n "github.com/mylukin/easy-i18n/i18n"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var CreateProjectCommand = cli.Command{
	Name:      "create",
	Aliases:   []string{"c"},
	Usage:     ei18n.Sprintf("create a project"),
	ArgsUsage: `[project name]`,
	Action: func(c *cli.Context) error {
		projectName := c.Args().Get(0)
		if projectName == "" {
			return errors.New(ei18n.Sprintf(`[project name] can't be empty.`))
		}
		return createProject(projectName)
	},
}

// TEMPLATE_URL 模板地址
const TEMPLATE_URL = "https://github.com/mylukin/EchoPilot-Template/archive/refs/heads/main.zip"

// ExecuteCmd1 执行命令逻辑
func createProject(packageName string) error {

	// 检查packageName 必须是这种格式 mylukin/example，否则报错
	if !strings.Contains(packageName, "/") {
		return errors.New("package name must be in the format of 'mylukin/example'")
	}

	// mylukin/app 生成 app, 根据 / 分割 取最后一个字符串
	projectName := filepath.Base(packageName)
	log.Println("project name:", projectName)

	// 下载模板
	log.Println("downloading template...")
	err := downloadAndUnzip(TEMPLATE_URL, projectName)
	if err != nil {
		panic(err)
	}

	log.Println("replace template...")
	projectTitle := cases.Title(language.English).String(projectName)
	// 替换 包名 github.com/mylukin/EchoPilot-Template
	replaceInFiles(projectName, []string{
		"github.com/mylukin/EchoPilot-Template",
		"EchoPilot-Template",
		"{APP_NAME}",
		"{APP_NAME_LOWER}",
		"{PACKAGE_NAME}",
	}, []string{
		"github.com/" + packageName,
		projectTitle,
		projectTitle,
		strings.ToLower(projectTitle),
		strings.ToLower(packageName),
	})

	// cp .env.example .env
	log.Println("copy .env.example to .env")
	os.Rename(projectName+"/.env.example", projectName+"/.env")

	// 执行 go mod tidy & go mod vendor
	log.Println("installing dependencies...")
	cmd := exec.Command("sh", "-c", `cd `+projectName+` && go mod tidy && go mod vendor`)

	// 运行命令，并获取其输出
	_, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		log.Println("exec sh error:", cmdErr)
		return nil
	}

	// 输出安装完成
	log.Println("install done!")

	return nil
}

func downloadAndUnzip(url, dest string) error {
	// 下载zip文件
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "template-*.zip")
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// 将下载的内容写入临时文件
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}

	// 解压缩
	return unzip(tmpFile.Name(), dest)
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(dest, 0755)

	for _, f := range r.File {
		// 调整文件路径以去除顶层目录
		path := adjustPath(f.Name)
		if path == "" {
			continue // 跳过根目录本身
		}
		fullPath := dest + "/" + path

		if f.FileInfo().IsDir() {
			os.MkdirAll(fullPath, f.Mode())
			continue
		}

		dirPath := filepath.Dir(fullPath) // 安全获取目录路径
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		zippedFile, err := f.Open()
		if err != nil {
			outFile.Close() // 尝试关闭文件，避免资源泄漏
			return err
		}

		_, err = io.Copy(outFile, zippedFile)

		outFile.Close()    // 关闭文件
		zippedFile.Close() // 关闭zip文件中的文件

		if err != nil {
			return err
		}
	}
	return nil
}

// adjustPath 移除路径中的顶层目录
func adjustPath(filePath string) string {
	// 分割路径
	parts := strings.SplitN(filePath, "/", 2)
	if len(parts) < 2 {
		return "" // 如果没有子目录或文件，返回空字符串
	}
	return parts[1] // 返回去除了顶层目录后的路径
}

// replaceInFiles 批量遍历指定目录下的所有文件，并替换文件内容
func replaceInFiles(rootDir string, oldStrings, newStrings []string) error {
	// 确保替换字符串的数组长度相同
	if len(oldStrings) != len(newStrings) {
		return errors.New("oldStrings and newStrings must have the same length")
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			err := replaceInFile(path, oldStrings, newStrings)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// replaceInFile 批量替换文件中的字符串
func replaceInFile(filePath string, oldStrings, newStrings []string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(fileContent)
	// 对于每一对 oldString 和 newString，执行替换
	for i, oldString := range oldStrings {
		newString := newStrings[i]
		content = strings.Replace(content, oldString, newString, -1)
	}

	// 如果内容未发生变化，则不需要重写文件
	if content == string(fileContent) {
		return nil
	}

	// 写回文件
	err = os.WriteFile(filePath, []byte(content), 0666)
	if err != nil {
		return err
	}

	return nil
}
