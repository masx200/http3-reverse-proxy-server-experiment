package load_balance

import "net/http"

func Get_List_Of_HTTP_Idempotent_Methods() []string {

	// Get_List_Of_HTTP_Idempotent_Methods 返回一个包含所有HTTP幂等方法名称的字符串切片

	return []string{"GET", "PUT", "DELETE", "HEAD", "OPTIONS"}
}

//在Go语言中，您可以编写如下函数来判断给定的`*http.Request`对象的Method是否为HTTP幂等方法：

// IsIdempotentMethod 判断给定的http.Request的Method是否为HTTP幂等方法，如果是则返回true，否则返回false
func IsIdempotentMethodFailoverAttemptStrategy(req *http.Request) bool {
	idempotentMethods := map[string]bool{
		http.MethodGet:     true,
		http.MethodPut:     true,
		http.MethodDelete:  true,
		http.MethodHead:    true,
		http.MethodOptions: true,
	}

	return idempotentMethods[req.Method]
}

//在这个函数`IsIdempotentMethod`中，我们创建了一个名为`idempotentMethods`的映射（map），其中键为HTTP幂等方法名，值为`true`。然后通过检查传入的`*http.Request`对象的`Method`属性是否存在于该映射中，来判断该请求方法是否为幂等方法。如果存在，则返回`true`，否则返回`false`。
