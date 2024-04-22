package load_balance

func Get_List_Of_HTTP_Idempotent_Methods() []string {

	// Get_List_Of_HTTP_Idempotent_Methods 返回一个包含所有HTTP幂等方法名称的字符串切片

	return []string{"GET", "PUT", "DELETE", "HEAD", "OPTIONS"}
}
