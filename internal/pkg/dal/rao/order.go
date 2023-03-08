package rao

type GetTeamBuyVersionReq struct {
	TeamID string `json:"team_id"`
}

type TeamBuyVersionResp struct {
	TeamBuyVersionType int32                  `json:"team_buy_version_type"`
	Title              string                 `json:"title"`
	UnitPriceExplain   string                 `json:"unit_price_explain"`
	MinUserNum         int64                  `json:"min_user_num"`
	MaxUserNum         int64                  `json:"max_user_num"`
	UsableMaxUserNum   int64                  `json:"usable_max_user_num"`
	ExistUserNum       int                    `json:"exist_user_num"`
	ExpirationDate     int64                  `json:"expiration_date"`
	Detail             []TeamBuyVersionDetail `json:"detail"`
}

type TeamBuyVersionDetail struct {
	Icon    TeamBuyVersionIcon `json:"icon"`
	Name    string             `json:"name"`
	Explain string             `json:"explain"`
}

type TeamBuyVersionIcon struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
}

type VumBuyVersionResp struct {
	Title  string                 `json:"title"`
	Detail []TeamBuyVersionDetail `json:"detail"`
}

type VumBuyVersionList struct {
	ID        int32   `json:"id"`
	Title     string  `json:"title"`      // 购买套餐名称
	VumCount  int64   `json:"vum_count"`  // VUM额度
	UnitPrice float64 `json:"unit_price"` // 单价
	Discounts float64 `json:"discounts"`
}

type TeamTrialExpirationDateResp struct {
	TrialExpirationDate   int64 `json:"trial_expiration_date"`
	TrialExpirationDayNum int64 `json:"trial_expiration_day_num"`
}

type GetOrderAmountDetailReq struct {
	CardType    int    `json:"card_type"`
	UserNum     int    `json:"user_num"`
	MonthNum    int    `json:"month_num"`
	OrderType   int32  `json:"order_type"`
	VumBuyNum   int    `json:"vum_buy_num"`
	VumCardType int32  `json:"vum_card_type"`
	TeamID      string `json:"team_id"`
}

type GetOrderAmountDetailResp struct {
	Amount                     float64  `json:"amount"`
	TeamBugVersionAmountDetail []string `json:"team_bug_version_amount_detail"`
	Discounts                  float64  `json:"discounts"`
	CostDetail                 string   `json:"cost_detail"`
}

type GetOrderListReq struct {
	TeamID string `json:"team_id"`

	Page int   `json:"page" form:"page,default=1"`
	Size int   `json:"size" form:"size,default=10"`
	Sort int32 `json:"sort" form:"sort"`
}

type GetNotPayOrderListReq struct {
	TeamID string `json:"team_id"`
}

type GetOrderListResp struct {
	OrderList []*OrderListResp `json:"order_list"`
	Total     int64            `json:"total"`
}

type OrderListResp struct {
	OrderID            string  `json:"order_id"`        // 订单号id
	OrderType          int32   `json:"order_type"`      // 订单类型：1-成员续费，2-VUM资源包
	ProductType        int32   `json:"product_type"`    // 产品类型：1-月付，2-年付，3-体验版，4-基础版，5-专业版
	VumNum             int32   `json:"vum_num"`         // VUM额度
	BuyNumber          int64   `json:"buy_number"`      // 购买数量（团队人员数量/VUM资源包数量）
	OrderAmount        float64 `json:"order_amount"`    // 订单金额
	PayType            int32   `json:"pay_type"`        // 支付方式：1-微信，2-支付宝，3-银联，4-PayPal
	PayStatus          int32   `json:"pay_status"`      // 支付状态：0-待支付，1-已支付，2-退款中，3-已退款
	FinishPayTime      int64   `json:"finish_pay_time"` // 到账时间
	PayUserName        string  `json:"pay_user_name"`   // 支付人名称
	GoodsValidDate     int32   `json:"goods_valid_date"`
	VumBuyVersionType  int32   `json:"vum_buy_version_type"`
	TeamBuyVersionType int32   `json:"team_buy_version_type"`
	OpenInvoiceState   int32   `json:"open_invoice_state"`
	CreatedAt          int64   `json:"created_at"` // 创建时间
}

type BatchDeleteOrderReq struct {
	TeamID   string   `json:"team_id"`
	OrderIDs []string `json:"order_ids"`
}

type GetInvoiceListReq struct {
	TeamID string `json:"team_id"`

	Page int   `json:"page" form:"page,default=1"`
	Size int   `json:"size" form:"size,default=10"`
	Sort int32 `json:"sort" form:"sort"`
}

type GetInvoiceListResp struct {
	InvoiceList []*InvoiceListResp `json:"invoice_list"`
	Total       int64              `json:"total"`
}

type InvoiceListResp struct {
	ID                 int64   `json:"id"`
	OrderID            string  `json:"order_id"` // 订单id
	TeamID             string  `json:"team_id"`  // 团队id
	OrderType          int32   `json:"order_type"`
	ProductType        int32   `json:"product_type"`       // 产品类型：1-月付，2-年付，3-体验版，4-基础版，5-专业版
	BuyNumber          int64   `json:"buy_number"`         // 购买数量（团队人员数量/VUM资源包数量）
	OrderAmount        float64 `json:"order_amount"`       // 订单金额
	InvoiceTitle       string  `json:"invoice_title"`      // 发票抬头
	InvoiceType        int32   `json:"invoice_type"`       // 发票类型：1-普通发票，2-专业发票
	ReceiveEmail       string  `json:"receive_email"`      // 接受邮箱
	CreatedAt          int64   `json:"created_at"`         // 申请时间
	OpenInvoiceState   int32   `json:"open_invoice_state"` // 申请状态：0-待开票，1-已开票，2-已作废
	VumBuyVersionType  int32   `json:"vum_buy_version_type"`
	TeamBuyVersionType int32   `json:"team_buy_version_type"`
}

type GetInvoiceDetailReq struct {
	TeamID string `json:"team_id"`
	ID     int64  `json:"id"`
}

type InvoiceDetail struct {
	ID              int64   `json:"id"`               // 主键id
	InvoiceTitle    string  `json:"invoice_title"`    // 发票抬头
	InvoiceType     int32   `json:"invoice_type"`     // 发票类型：1-普通发票，2-专业发票
	TaxNum          string  `json:"tax_num"`          // 纳税识别号
	CompanyAddress  string  `json:"company_address"`  // 公司地址（开票地址）
	OpenBankName    string  `json:"open_bank_name"`   // 开户银行名称
	BankAccountNum  string  `json:"bank_account_num"` // 开户行账号
	Phone           string  `json:"phone"`            // 电话号码
	OpenInvoiceMode int32   `json:"open_invoice_mode"`
	ReceiveEmail    string  `json:"receive_email"` // 接受邮箱
	ReceiverName    string  `json:"receiver_name"`
	ReceiverPhone   string  `json:"receiver_phone"`
	ReceiverAddress string  `json:"receiver_address"`
	RealAmount      float64 `json:"real_amount"`
}

type GetInvoiceDetailResp struct {
	InvoiceDetail InvoiceDetail `json:"invoice_detail"`
}

type AddInvoiceReq struct {
	TeamID          string `json:"team_id"`
	OrderID         string `json:"order_id"`
	InvoiceTitle    string `json:"invoice_title"`     // 发票抬头
	InvoiceType     int32  `json:"invoice_type"`      // 发票类型：1-普通发票，2-专业发票
	TaxNum          string `json:"tax_num"`           // 纳税识别号
	CompanyAddress  string `json:"company_address"`   // 公司地址（开票地址）
	OpenBankName    string `json:"open_bank_name"`    // 开户银行名称
	BankAccountNum  string `json:"bank_account_num"`  // 开户行账号
	Phone           string `json:"phone"`             // 电话号码
	ReceiveEmail    string `json:"receive_email"`     // 接受邮箱
	OpenInvoiceMode int32  `json:"open_invoice_mode"` // 开票方式：1-专票-电子票，2-专票-邮寄
	ReceiverName    string `json:"receiver_name"`     // 收件人姓名
	ReceiverPhone   string `json:"receiver_phone"`    // 收件人电话
	ReceiverAddress string `json:"receiver_address"`  // 收件人地址
}

type GetOrderDetailReq struct {
	OrderID string `json:"order_id"`
	TeamID  string `json:"team_id"`
}

type OrderDetail struct {
	OrderID          string  `json:"order_id"`           // 订单号id
	OrderType        int32   `json:"order_type"`         // 订单类型：1-成员续费，2-VUM资源包
	ProductType      int32   `json:"product_type"`       // 产品类型：1-月付，2-年付，3-体验版，4-基础版，5-专业版
	VumNum           int64   `json:"vum_num"`            // VUM额度
	BuyNumber        int64   `json:"buy_number"`         // 购买数量（团队人员数量/VUM资源包数量）
	OrderAmount      float64 `json:"order_amount"`       // 订单金额
	RealAmount       float64 `json:"real_amount"`        // 实际付款金额
	Discounts        float64 `json:"discounts"`          // 优惠金额
	PayType          int32   `json:"pay_type"`           // 支付方式：1-微信，2-支付宝，3-银联，4-PayPal
	PayStatus        int32   `json:"pay_status"`         // 支付状态：0-待支付，1-已支付，2-退款中，3-已退款
	GoodsValidDate   string  `json:"goods_valid_date"`   // 商品的有效期（单位：月）
	FinishPayTime    int64   `json:"finish_pay_time"`    // 到账时间
	PayUserName      string  `json:"pay_user_name"`      // 支付人名称
	CreatedAt        int64   `json:"created_at"`         // 创建时间
	TotalVumNum      int64   `json:"total_vum_num"`      // VUM总额度
	OpenInvoiceState int32   `json:"open_invoice_state"` // 发票申请状态
}

type GetOrderDetailResp struct {
	OrderDetail *OrderDetail `json:"order_detail"`
}

type GetOrderPayStatusReq struct {
	OrderID string `json:"order_id"`
}

type GetOrderPayStatusResp struct {
	OrderPayStatus int32  `json:"order_pay_status"`
	TeamID         string `json:"team_id"`
	OverdueTime    string `json:"overdue_time"`
	OrderEnd       bool   `json:"order_end"`
}

type GetVumUseListReq struct {
	TeamID string `json:"team_id"`

	Page int   `json:"page" form:"page,default=1"`
	Size int   `json:"size" form:"size,default=10"`
	Sort int32 `json:"sort" form:"sort"`
}

type GetVumUseListResp struct {
	VumUseList   []*VumUseListResp `json:"vum_use_list"`
	Total        int64             `json:"total"`
	UsedVumNum   int64             `json:"used_vum_num"`
	UsableVumNum int64             `json:"usable_vum_num"`
}

type VumUseListResp struct {
	ID                int64  `json:"id"`
	PlanID            string `json:"plan_id"`            // 计划id
	PlanName          string `json:"plan_name"`          // 计划名称
	TaskType          int32  `json:"task_type"`          // 任务类型：1-普通任务，2-定时任务
	TaskMode          int32  `json:"task_mode"`          // 压测模式：1-并发模式，2-阶梯模式，3-错误率模式，4-响应时间模式，5-每秒请求数模式，6 -每秒事务数模式
	RunTime           int64  `json:"run_time"`           // 运行时间
	RunUserName       string `json:"run_user_id"`        // 执行者id
	ConcurrenceNum    int32  `json:"concurrence_num"`    // 并发数
	ConcurrenceMinute int32  `json:"concurrence_minute"` // 并发时长（单位分钟）
	VumConsumeNum     int64  `json:"vum_consume_num"`    // VUM使用量
}

type CreateOrderReq struct {
	TeamID    string `json:"team_id"`
	TeamName  string `json:"team_name"`
	OrderType int32  `json:"order_type"` // 订单类型：订单类型：1-新建团队，2-VUM资源包，3-升级团队，4-增加席位，5-套餐续期
	//ProductType       int32  `json:"product_type"` // 产品类型：1-个人版，2-团队版，3-企业版
	VumBuyVersionType int32 `json:"vum_buy_version_type"`
	BuyNumber         int64 `json:"buy_number"` // 购买数量（团队人员数量/VUM资源包数量）
	//OrderAmount        float64 `json:"order_amount"`          // 订单金额
	GoodsValidDate     int32 `json:"goods_valid_date"`      // 商品的有效期（单位：月）
	TeamBuyVersionType int32 `json:"team_buy_version_type"` // 团队套餐类型
}

type CreateOrderResp struct {
	OrderID string `json:"order_id"`
}

type GetQrCodeParam struct {
	OrderID     string  `json:"order_id"`
	TotalAmount float64 `json:"total_amount"`
	Business    string  `json:"business"`
	Subject     string  `json:"subject"`
	NotifyUrl   string  `json:"notify_url"`
	Theme       string  `json:"theme"`
}

type GetQrCodeResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		PayTradeNo string `json:"pay_trade_no"`
		ScanUrl    string `json:"scan_url"`
	} `json:"data"`
}

type GetPayResultParam struct {
	OrderID  string `json:"order_id"`
	Business string `json:"business"`
}

type GetPayResultResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Status      int32  `json:"status"`
		PaidAt      int    `json:"paid_at"`
		ClosedAt    int    `json:"closed_at"`
		TradeNo     string `json:"trade_no"`
		OrderId     string `json:"order_id"`
		Business    string `json:"business"`
		PayTradeNo  string `json:"pay_trade_no"`
		TotalAmount string `json:"total_amount"`
		PayMethod   string `json:"pay_method"`
	} `json:"data"`
}

type PayNotifyApiReq struct {
	Status      int32  `json:"status"`
	PaidAt      int64  `json:"paid_at"`
	ClosedAt    int64  `json:"closed_at"`
	TradeNo     string `json:"trade_no"`
	OrderId     string `json:"order_id"`
	Business    string `json:"business"`
	PayTradeNo  string `json:"pay_trade_no"`
	TotalAmount string `json:"total_amount"`
	PayMethod   string `json:"pay_method"`
}

type GetVumAmountReq struct {
	CardType int32 `json:"card_type"` // 资源包类型：1-套餐A，2-套餐B，3-套餐C，4-套餐D
	VumNum   int64 `json:"vum_num"`   // VUM额度
	BuyNum   int   `json:"buy_num"`   // 购买数量
}

type GetVumAmountResp struct {
	Amount          float64  `json:"amount"`
	VumAmountDetail []string `json:"vum_amount_detail"`
}

type GetMyResourceInfoReq struct {
	TeamID string `json:"team_id"`
}

type GetMyResourceInfoResp struct {
	TotalVumNum        int64                    `json:"total_vum_num"`
	UsableVumNum       int64                    `json:"usable_vum_num"`
	VumPercent         float64                  `json:"vum_percent"`
	UsedVumNum         int64                    `json:"used_vum_num"`
	TeamBuyVersionType int32                    `json:"team_buy_version_type"`
	MaxUserNum         int64                    `json:"max_user_num"`
	CurrentTeamUserNum int64                    `json:"current_team_user_num"`
	VipExpirationDate  int64                    `json:"vip_expiration_date"`
	CanAddUserNum      int64                    `json:"can_add_user_num"`
	TeamBuyVersionInfo [][]TeamBuyVersionDetail `json:"team_buy_version_info"`
}

type GetOrderPayDetailReq struct {
	OrderID string `json:"order_id"`
}

type OrderPayDetail struct {
	OrderID   string `json:"order_id"` // 订单号id
	TeamName  string `json:"team_name"`
	OrderType int32  `json:"order_type"` // 订单类型：1-成员续费，2-VUM资源包
	//ProductType         int32                  `json:"product_type"`  // 产品类型：1-月付，2-年付，3-体验版，4-基础版，5-专业版
	TeamBuyVersionType    int32                  `json:"team_buy_version_type"`
	NewTeamBuyVersionType int32                  `json:"new_team_buy_version_type"`
	TotalVumNum           int64                  `json:"total_vum_num"` // VUM额度
	BuyNum                int64                  `json:"buy_num"`
	OrderAmount           float64                `json:"order_amount"`           // 订单金额
	Discounts             float64                `json:"discounts"`              // 优惠金额
	GoodsValidDate        int32                  `json:"goods_valid_date"`       // 商品的有效期（单位：月）
	GoodsValidDateStart   int64                  `json:"goods_valid_date_start"` // 商品的有效期-开始时间
	GoodsValidDateEnd     int64                  `json:"goods_valid_date_end"`   // 商品的有效期-结束时间
	CreatedAt             int64                  `json:"created_at"`             // 创建时间
	ScanUrl               string                 `json:"scan_url"`
	TeamBuyVersionInfo    []TeamBuyVersionDetail `json:"team_buy_version_info"`
	TeamMaxUserNum        int64                  `json:"team_max_user_num"`
	BillingMethod         string                 `json:"billing_method"`
	AmountDetail          string                 `json:"amount_detail"`
	VumBuyVersionType     int32                  `json:"vum_buy_version_type"`
	VumCount              int64                  `json:"vum_count"`
}

type GetOrderPayDetailResp struct {
	OrderPayDetail *OrderPayDetail `json:"order_pay_detail"`
}

type GetCurrentTeamBuyVersionReq struct {
	TeamID string `json:"team_id"`
}

type GetVumBuyVersionReq struct {
	TeamID string `json:"team_id"`
}
type GetVumBuyVersionResp struct {
	TeamName          string              `json:"team_name"`
	MaxConcurrence    int64               `json:"max_concurrence"`
	VumBuyVersionList []VumBuyVersionList `json:"vum_buy_version_list"`
}

type GetCurrentTeamBuyVersionResp struct {
	MaxConcurrence     int64  `json:"max_concurrence"`       // 最大并发数
	MaxAPINum          int64  `json:"max_api_num"`           // 最大接口数
	MaxRunTime         int64  `json:"max_run_time"`          // 最大运行时长
	TeamBuyVersionType int32  `json:"team_buy_version_type"` // 团队套餐类型
	MaxUserNum         int64  `json:"max_user_num"`          // 最大成员数量
	ExistUserNum       int    `json:"exist_user_num"`        // 当前团队已存在用户数量
	TeamName           string `json:"team_name"`             // 团队名称
	ExpirationDate     int64  `json:"expiration_date"`       // 截止日期
	TeamMaxUserNum     int64  `json:"team_max_user_num"`     // 团队已买席位数
}
