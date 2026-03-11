package main

import (
	"bytes"
	"context"
	"encoding/json"
	"go/format"
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

// --- 结构体定义 ---

type RunRequest struct {
	Code   string `json:"code"`
	Input  string `json:"input"`
	TaskID string `json:"taskId"`
}

type RunResponse struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"isDir"`
	Children []*FileNode `json:"children,omitempty"`
}

type FormatRequest struct {
	Code string `json:"code"`
}

type FormatResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

// --- 核心逻辑重构 ---

// executeTask 负责编译和运行的具体生命周期管理
func executeTask(ctx context.Context, req RunRequest, tempDir string) RunResponse {
	codePath := filepath.Join(tempDir, "main.go")
	binPath := filepath.Join(tempDir, "prog")

	if err := os.WriteFile(codePath, []byte(req.Code), 0644); err != nil {
		return RunResponse{Error: "Internal Server Error: Failed to write code"}
	}

	// 1. 编译阶段
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, codePath)
	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr

	if err := buildCmd.Run(); err != nil {
		if ctx.Err() != nil {
			return RunResponse{} // 上下文取消，不返回错误
		}
		return RunResponse{Error: "Compilation Error:\n" + buildStderr.String()}
	}

	// 2. 运行阶段
	return runBinary(ctx, binPath, req.Input)
}

// runBinary 处理二进制文件的执行过程
func runBinary(ctx context.Context, binPath string, input string) RunResponse {
	cmd := exec.CommandContext(ctx, binPath)
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err := cmd.Start(); err != nil {
		return RunResponse{Error: "Start Error: " + err.Error()}
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-ctx.Done():
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return RunResponse{Error: "Program stopped (Manual/Timeout)"}
	case err := <-done:
		resp := RunResponse{Output: stdout.String()}
		if err != nil {
			resp.Error = stderr.String()
			if resp.Error == "" {
				resp.Error = err.Error()
			}
		}
		return resp
	}
}

// --- HTTP 处理函数 ---

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

	tempDir, err := os.MkdirTemp("", "go-run-")
	if err != nil {
		http.Error(w, "Failed to create temp dir", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	// 绑定请求上下文，处理连接断开和 1 分钟超时
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Minute)
	defer cancel()

	if req.TaskID != "" {
		runningTasks.Store(req.TaskID, cancel)
		defer runningTasks.Delete(req.TaskID)
	}

	resp := executeTask(ctx, req, tempDir)

	// 如果是因为连接断开导致的取消，无需返回数据给已经关闭的连接
	if r.Context().Err() != nil {
		log.Printf("Task %s cleaned up due to client disconnection", req.TaskID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func stopCodeHandler(w http.ResponseWriter, r *http.Request) {
	taskId := r.URL.Query().Get("taskId")
	if taskId == "" {
		http.Error(w, "Missing taskId", http.StatusBadRequest)
		return
	}

	if cancelFunc, ok := runningTasks.Load(taskId); ok {
		cancelFunc.(context.CancelFunc)()
		runningTasks.Delete(taskId)
		w.Write([]byte("OK"))
	} else {
		http.Error(w, "Task not found or already finished", http.StatusNotFound)
	}
}

func formatCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	formatted, err := format.Source([]byte(req.Code))
	var resp FormatResponse
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Code = string(formatted)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// --- 文件系统处理 ---

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
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			childNode, err := buildFileTree(filepath.Join(dir, entry.Name()))
			if err == nil && childNode != nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}
	return node, nil
}

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

func getTemplateContentHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	cleanPath := filepath.Clean(path)

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

// --- 主函数 ---

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/run", runCodeHandler)
	http.HandleFunc("/api/stop", stopCodeHandler)
	http.HandleFunc("/api/format", formatCodeHandler)
	http.HandleFunc("/api/templates", getTemplatesHandler)
	http.HandleFunc("/api/template/content", getTemplateContentHandler)

	log.Println("Server is running on http://0.0.0.0:2009")
	if err := http.ListenAndServe(":2009", nil); err != nil {
		log.Fatal(err)
	}
}
