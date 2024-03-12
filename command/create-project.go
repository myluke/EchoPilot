package command

import (
	"archive/zip"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	ei18n "github.com/mylukin/easy-i18n/i18n"
	"github.com/urfave/cli/v2"
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
func createProject(name string) error {
	err := downloadAndUnzip(TEMPLATE_URL, name)
	if err != nil {
		panic(err)
	}
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
