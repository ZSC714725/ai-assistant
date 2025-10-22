package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
)

// Config 配置结构体
type Config struct {
	API struct {
		BaseURL string `yaml:"base_url"`
		APIKey  string `yaml:"api_key"`
	} `yaml:"api"`
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Models struct {
		Default   string   `yaml:"default"`
		Available []string `yaml:"available"`
	} `yaml:"models"`
}

// ChatRequest 聊天请求结构体
type ChatRequest struct {
	Message string `json:"message" binding:"required"`
	Model   string `json:"model"`
}

// ChatResponse 聊天响应结构体
type ChatResponse struct {
	Response string `json:"response"`
	Model    string `json:"model"`
}

// QARecord 问答记录结构体
type QARecord struct {
	ID        int       `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Model     string    `json:"model"`
	Timestamp time.Time `json:"timestamp"`
}

// KnowledgeItem 知识库条目结构体
type KnowledgeItem struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Model     string    `json:"model"`
	Timestamp time.Time `json:"timestamp"`
	Tags      []string  `json:"tags"`
}

// AddToKnowledgeRequest 添加到知识库请求
type AddToKnowledgeRequest struct {
	RecordID int    `json:"record_id" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Tags     string `json:"tags"`
}

var config Config
var recentQAs []QARecord
var knowledgeBase []KnowledgeItem
var nextQAID = 1
var nextKnowledgeID = 1

// 数据文件路径
const (
	knowledgeDataFile = "data/knowledge.json"
	qaDataFile        = "data/recent_qas.json"
)

func main() {
	// 加载配置文件
	loadConfig()

	// 加载持久化数据
	loadPersistentData()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由
	r := gin.Default()

	// 静态文件服务
	r.Static("/static", "./static")

	// 主页路由
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./templates/index.html")
	})

	// index.html 路由
	r.GET("/index.html", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./templates/index.html")
	})

	// API路由
	api := r.Group("/api")
	{
		api.POST("/chat", chatHandler)
		api.GET("/models", modelsHandler)
		api.GET("/recent", recentQAsHandler)
		api.POST("/knowledge/add", addToKnowledgeHandler)
		api.GET("/knowledge", knowledgeHandler)
		api.DELETE("/knowledge/:id", deleteKnowledgeHandler)
	}

	// 知识库页面路由
	r.GET("/knowledge", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./templates/knowledge.html")
	})

	// 启动服务器
	address := config.Server.Host + config.Server.Port
	fmt.Printf("服务器启动在: http://%s\n", address)
	log.Fatal(r.Run(address))
}

// loadConfig 加载配置文件
func loadConfig() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
}

// chatHandler 处理聊天请求
func chatHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果没有指定模型，使用默认模型
	if req.Model == "" {
		req.Model = config.Models.Default
	}

	// 调用OpenAI API
	response, err := callWithOfficialSDK(req.Message, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 记录问答到最近记录
	record := QARecord{
		ID:        nextQAID,
		Question:  req.Message,
		Answer:    response,
		Model:     req.Model,
		Timestamp: time.Now(),
	}

	// 添加到最近记录，保持最多5条
	recentQAs = append([]QARecord{record}, recentQAs...)
	if len(recentQAs) > 5 {
		recentQAs = recentQAs[:5]
	}
	nextQAID++

	// 保存问答数据到文件
	saveRecentQAs()

	c.JSON(http.StatusOK, ChatResponse{
		Response: response,
		Model:    req.Model,
	})
}

// modelsHandler 返回可用模型列表
func modelsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"default":   config.Models.Default,
		"available": config.Models.Available,
	})
}

// recentQAsHandler 返回最近5次问答记录
func recentQAsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"recent_qas": recentQAs,
	})
}

// addToKnowledgeHandler 将问答记录添加到知识库
func addToKnowledgeHandler(c *gin.Context) {
	var req AddToKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找对应的问答记录
	var sourceRecord *QARecord
	for _, record := range recentQAs {
		if record.ID == req.RecordID {
			sourceRecord = &record
			break
		}
	}

	if sourceRecord == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到对应的问答记录"})
		return
	}

	// 解析标签
	var tags []string
	if req.Tags != "" {
		tags = strings.Split(req.Tags, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	// 创建知识库条目
	knowledgeItem := KnowledgeItem{
		ID:        nextKnowledgeID,
		Title:     req.Title,
		Content:   sourceRecord.Answer,
		Model:     sourceRecord.Model,
		Timestamp: time.Now(),
		Tags:      tags,
	}

	knowledgeBase = append(knowledgeBase, knowledgeItem)
	nextKnowledgeID++

	// 保存知识库数据到文件
	saveKnowledgeBase()

	c.JSON(http.StatusOK, gin.H{
		"message": "已成功添加到知识库",
		"item":    knowledgeItem,
	})
}

// knowledgeHandler 返回知识库内容
func knowledgeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"knowledge_base": knowledgeBase,
	})
}

// deleteKnowledgeHandler 删除知识库条目
func deleteKnowledgeHandler(c *gin.Context) {
	id := c.Param("id")

	// 简单的字符串转整数（实际项目中应该使用strconv.Atoi）
	var targetID int
	fmt.Sscanf(id, "%d", &targetID)

	// 查找并删除
	for i, item := range knowledgeBase {
		if item.ID == targetID {
			knowledgeBase = append(knowledgeBase[:i], knowledgeBase[i+1:]...)

			// 保存知识库数据到文件
			saveKnowledgeBase()

			c.JSON(http.StatusOK, gin.H{"message": "已删除知识库条目"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "未找到对应的知识库条目"})
}

func callWithOfficialSDK(content, model string) (string, error) {
	openaiConfig := openai.DefaultConfig(config.API.APIKey)
	openaiConfig.BaseURL = config.API.BaseURL

	client := openai.NewClientWithConfig(openaiConfig)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	// 返回AI响应内容
	return resp.Choices[0].Message.Content, nil
}

// loadPersistentData 加载持久化数据
func loadPersistentData() {
	// 确保data目录存在
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Printf("创建data目录失败: %v", err)
	}

	// 加载知识库数据
	loadKnowledgeBase()

	// 加载最近问答数据
	loadRecentQAs()
}

// loadKnowledgeBase 加载知识库数据
func loadKnowledgeBase() {
	if _, err := os.Stat(knowledgeDataFile); os.IsNotExist(err) {
		// 文件不存在，使用空数据
		knowledgeBase = []KnowledgeItem{}
		return
	}

	data, err := ioutil.ReadFile(knowledgeDataFile)
	if err != nil {
		log.Printf("读取知识库数据失败: %v", err)
		knowledgeBase = []KnowledgeItem{}
		return
	}

	var items []KnowledgeItem
	if err := json.Unmarshal(data, &items); err != nil {
		log.Printf("解析知识库数据失败: %v", err)
		knowledgeBase = []KnowledgeItem{}
		return
	}

	knowledgeBase = items

	// 更新下一个ID
	if len(knowledgeBase) > 0 {
		maxID := 0
		for _, item := range knowledgeBase {
			if item.ID > maxID {
				maxID = item.ID
			}
		}
		nextKnowledgeID = maxID + 1
	}

	log.Printf("已加载 %d 条知识库记录", len(knowledgeBase))
}

// loadRecentQAs 加载最近问答数据
func loadRecentQAs() {
	if _, err := os.Stat(qaDataFile); os.IsNotExist(err) {
		// 文件不存在，使用空数据
		recentQAs = []QARecord{}
		return
	}

	data, err := ioutil.ReadFile(qaDataFile)
	if err != nil {
		log.Printf("读取问答数据失败: %v", err)
		recentQAs = []QARecord{}
		return
	}

	var qas []QARecord
	if err := json.Unmarshal(data, &qas); err != nil {
		log.Printf("解析问答数据失败: %v", err)
		recentQAs = []QARecord{}
		return
	}

	recentQAs = qas

	// 更新下一个ID
	if len(recentQAs) > 0 {
		maxID := 0
		for _, qa := range recentQAs {
			if qa.ID > maxID {
				maxID = qa.ID
			}
		}
		nextQAID = maxID + 1
	}

	log.Printf("已加载 %d 条问答记录", len(recentQAs))
}

// saveKnowledgeBase 保存知识库数据
func saveKnowledgeBase() {
	data, err := json.MarshalIndent(knowledgeBase, "", "  ")
	if err != nil {
		log.Printf("序列化知识库数据失败: %v", err)
		return
	}

	if err := ioutil.WriteFile(knowledgeDataFile, data, 0644); err != nil {
		log.Printf("保存知识库数据失败: %v", err)
	}
}

// saveRecentQAs 保存最近问答数据
func saveRecentQAs() {
	data, err := json.MarshalIndent(recentQAs, "", "  ")
	if err != nil {
		log.Printf("序列化问答数据失败: %v", err)
		return
	}

	if err := ioutil.WriteFile(qaDataFile, data, 0644); err != nil {
		log.Printf("保存问答数据失败: %v", err)
	}
}
