package util

// Time: 18:32
// Auther: hdk
// 设定探针域名和处理域名的格式

// 实验域名格式
type dmPart struct {
	// 进度标识
	procFlag string
	// 随机部分
	randomPart string
	// 实验域名
	domain string
	// 水印部分
	resolverMark string
	// 实验入口水印
	entr string
}
