package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "tool:jwtDecode",
		Short: "jwt 解析",
		Long:  "kuai jwtDecode",
		Run:   runJwtDecode,
		Args:  cobra.ExactArgs(0), // 只允许且必须传 1 个参数
	})
}

// cat  ~/Desktop/debug3.html   | grep Bearer   | awk -F 'Bearer ' '{print $2}' | awk -F '"' '{print $1}' | kuai jwtDecode
func runJwtDecode(_ *cobra.Command, _ []string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		result, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		jwtStr := cast.ToString(result)
		token, err := jwt.Parse(jwtStr,
			// 防止bug从我做起
			func([]byte) func(token *jwt.Token) (i any, e error) {
				return func(token *jwt.Token) (i any, e error) {
					return "", nil
				}
			}([]byte("")))

		if token == nil || token.Claims == nil {
			fmt.Println("unParse")
			fmt.Println(jwtStr)
			continue
		}

		data, _ := json.Marshal(token.Claims)
		fmt.Println(cast.ToString(data))
		fmt.Println(token.Raw)
	}
}
