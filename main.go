package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 任务管理器：存储正在运行的任务取消函数
var runningTasks sync.Map

// 请求体结构
type RunRequest struct {
	Code   string `json:"code"`
	Input  string `json:"input"`
	TaskID string `json:"taskId"`
}

// 响应体结构
type RunResponse struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

// 文件树节点结构
type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"isDir"`
	Children []*FileNode `json:"children,omitempty"`
}

func runCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 1. 准备工作目录
	tempDir, err := os.MkdirTemp("", "go-run-")
	if err != nil {
		http.Error(w, "Failed to create temp dir", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	codePath := filepath.Join(tempDir, "main.go")
	binPath := filepath.Join(tempDir, "prog")
	if err := os.WriteFile(codePath, []byte(req.Code), 0644); err != nil {
		http.Error(w, "Failed to write code", http.StatusInternalServerError)
		return
	}

	// 2. 【核心修改】绑定请求上下文
	// 如果用户刷新页面或断开连接，r.Context() 会取消，从而触发后续的清理逻辑
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	// 注册到任务管理器，支持手动停止
	if req.TaskID != "" {
		runningTasks.Store(req.TaskID, cancel)
		defer runningTasks.Delete(req.TaskID)
	}

	// 3. 编译阶段
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, codePath)
	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr
	if err := buildCmd.Run(); err != nil {
		if ctx.Err() != nil { // 说明是被取消或断开了
			return
		}
		json.NewEncoder(w).Encode(RunResponse{Error: "Compilation Error:\n" + buildStderr.String()})
		return
	}

	// 4. 运行阶段
	cmd := exec.CommandContext(ctx, binPath)

	if req.Input != "" {
		cmd.Stdin = strings.NewReader(req.Input)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		json.NewEncoder(w).Encode(RunResponse{Error: "Start Error: " + err.Error()})
		return
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	var resp RunResponse
	select {
	case <-ctx.Done():
		// 【清理逻辑】只要上下文结束（手动停止、连接断开、超时），就杀掉整个进程组
		if cmd.Process != nil {
			cmd.Process.Kill()
		}

		// 如果是因为连接断开导致的取消，无需返回数据给已经关闭的连接
		if r.Context().Err() != nil {
			log.Printf("Task %s cleaned up due to client disconnection", req.TaskID)
			return
		}

		resp.Error = "Program stopped (Manual/Timeout)"
	case err := <-done:
		resp.Output = stdout.String()
		if err != nil {
			resp.Error = stderr.String()
			if resp.Error == "" {
				resp.Error = err.Error()
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 构建文件树
func buildFileTree(dir string) (*FileNode, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	node := &FileNode{
		Name:  filepath.Base(dir),
		Path:  dir,
		IsDir: info.IsDir(),
	}
	if node.IsDir {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			// 过滤隐藏文件和空文件 (如 .git, .gitignore)
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			childPath := filepath.Join(dir, entry.Name())
			childNode, err := buildFileTree(childPath)
			if err == nil && childNode != nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}
	return node, nil
}

// 获取模板目录树接口
func getTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	root := "goacm"
	tree, err := buildFileTree(root)
	if err != nil {
		http.Error(w, "Failed to read goacm directory", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
}

// 获取具体模板文件内容接口
func getTemplateContentHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	cleanPath := filepath.Clean(path)

	// 路径安全检查，防止目录穿越读取其他文件
	if strings.Contains(cleanPath, "..") || !strings.HasPrefix(filepath.ToSlash(cleanPath), "goacm") {
		http.Error(w, "Invalid or forbidden path", http.StatusForbidden)
		return
	}

	content, err := os.ReadFile(cleanPath)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	w.Write(content)
}

// 停止接口
func stopCodeHandler(w http.ResponseWriter, r *http.Request) {
	taskId := r.URL.Query().Get("taskId")
	if taskId == "" {
		http.Error(w, "Missing taskId", http.StatusBadRequest)
		return
	}

	if cancelFunc, ok := runningTasks.Load(taskId); ok {
		cancelFunc.(context.CancelFunc)() // 执行 Context 的取消动作
		runningTasks.Delete(taskId)
		w.Write([]byte("OK"))
	} else {
		http.Error(w, "Task not found or already finished", http.StatusNotFound)
	}
}

func main() {
	// 静态文件服务
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API 路由
	http.HandleFunc("/api/run", runCodeHandler)
	http.HandleFunc("/api/stop", stopCodeHandler) // 新增停止接口
	http.HandleFunc("/api/templates", getTemplatesHandler)
	http.HandleFunc("/api/template/content", getTemplateContentHandler)

	log.Println("Server is running on http://0.0.0.0:2009")
	if err := http.ListenAndServe(":2009", nil); err != nil {
		log.Fatal(err)
	}
}
