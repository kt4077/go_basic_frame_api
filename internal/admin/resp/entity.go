package resp

import (
	"time"

	"gorm.io/gorm"

	"server_api/internal/common/model"
)

type BaseItem struct {
	ID        uint           `json:"id" comment:"主键ID"`
	CreatedAt time.Time      `json:"created_at" comment:"创建时间"`
	UpdatedAt time.Time      `json:"updated_at" comment:"更新时间"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index" comment:"删除时间"`
}

type DeptItem struct {
	BaseItem `comment:"基础字段"`
	Name     string `json:"name" comment:"部门名称"`
	ParentID uint   `json:"parent_id" comment:"上级部门ID"`
	Sort     int    `json:"sort" comment:"排序"`
	Leader   string `json:"leader" comment:"负责人"`
	Remark   string `json:"remark" comment:"备注"`
}

func (d DeptItem) GetID() uint       { return d.ID }
func (d DeptItem) GetParentID() uint { return d.ParentID }

type MenuItem struct {
	BaseItem `comment:"基础字段"`
	Name     string `json:"name" comment:"菜单名称"`
	Type     int    `json:"type" comment:"菜单类型"`
	ParentID uint   `json:"parent_id" comment:"上级菜单ID"`
	Path     string `json:"path" comment:"前端路由"`
	ApiPath  string `json:"api_path" comment:"接口权限地址"`
	Icon     string `json:"icon" comment:"图标"`
	Sort     int    `json:"sort" comment:"排序"`
	Status   int    `json:"status" comment:"状态"`
	Remark   string `json:"remark" comment:"备注"`
}

func (m MenuItem) GetID() uint       { return m.ID }
func (m MenuItem) GetParentID() uint { return m.ParentID }

type RoleItem struct {
	BaseItem `comment:"基础字段"`
	Name     string `json:"name" comment:"角色名称"`
	Code     string `json:"code" comment:"角色编码"`
	ParentID uint   `json:"parent_id" comment:"上级角色ID"`
	Sort     int    `json:"sort" comment:"排序"`
	Status   int    `json:"status" comment:"状态"`
	Remark   string `json:"remark" comment:"备注"`
}

func (r RoleItem) GetID() uint       { return r.ID }
func (r RoleItem) GetParentID() uint { return r.ParentID }

type UserItem struct {
	BaseItem `comment:"基础字段"`
	Username string     `json:"username" comment:"登录账号"`
	Nickname string     `json:"nickname" comment:"用户昵称"`
	Avatar   string     `json:"avatar" comment:"头像地址"`
	Mobile   string     `json:"mobile" comment:"手机号"`
	Email    string     `json:"email" comment:"邮箱"`
	DeptID   uint       `json:"dept_id" comment:"部门ID"`
	Status   int        `json:"status" comment:"状态"`
	IsSuper  int        `json:"is_super" comment:"是否超级管理员"`
	Roles    []RoleItem `json:"roles" gorm:"many2many:sys_user_role;joinForeignKey:UserID;joinReferences:RoleID" comment:"角色列表"`
}

// TableName 指定响应查询结构对应的用户表，使 GORM 可直接向 resp 预加载关联。
func (UserItem) TableName() string { return "sys_user" }

// TableName 指定响应查询结构对应的角色表。
func (RoleItem) TableName() string { return "sys_role" }

type StorageItem struct {
	BaseItem  `comment:"基础字段"`
	Name      string `json:"name" comment:"渠道名称"`
	Channel   string `json:"channel" comment:"渠道类型"`
	Params    string `json:"params" comment:"渠道参数JSON"`
	IsDefault int    `json:"is_default" comment:"是否默认"`
	Status    int    `json:"status" comment:"状态"`
	Sort      int    `json:"sort" comment:"排序"`
	Remark    string `json:"remark" comment:"备注"`
}

type SMSConfigItem struct {
	BaseItem    `comment:"基础字段"`
	Name        string `json:"name" comment:"配置名称"`
	Provider    int    `json:"provider" comment:"短信服务商"`
	AccessKeyID string `json:"access_key_id" comment:"访问密钥ID"`
	Endpoint    string `json:"endpoint" comment:"服务地址"`
	Status      int    `json:"status" comment:"状态"`
	Remark      string `json:"remark" comment:"备注"`
}
type SMSSignatureItem struct {
	BaseItem `comment:"基础字段"`
	ConfigID uint   `json:"config_id" comment:"短信配置ID"`
	Name     string `json:"name" comment:"签名名称"`
	SignCode string `json:"sign_code" comment:"平台签名编码"`
	Status   int    `json:"status" comment:"状态"`
	Remark   string `json:"remark" comment:"备注"`
}
type SMSTemplateItem struct {
	BaseItem     `comment:"基础字段"`
	ConfigID     uint   `json:"config_id" comment:"短信配置ID"`
	Name         string `json:"name" comment:"模板名称"`
	TemplateCode string `json:"template_code" comment:"平台模板编码"`
	Type         int    `json:"type" comment:"模板类型"`
	Content      string `json:"content" comment:"模板内容"`
	Status       int    `json:"status" comment:"状态"`
	Remark       string `json:"remark" comment:"备注"`
}
type SMSLogItem struct {
	ID                uint       `json:"id" comment:"主键ID"`
	CreatedAt         time.Time  `json:"created_at" comment:"创建时间"`
	ConfigID          uint       `json:"config_id" comment:"短信配置ID"`
	SignatureID       uint       `json:"signature_id" comment:"签名ID"`
	TemplateID        uint       `json:"template_id" comment:"模板ID"`
	Mobile            string     `json:"mobile" comment:"手机号"`
	Content           string     `json:"content" comment:"发送内容"`
	Status            int        `json:"status" comment:"发送状态"`
	ProviderMessageID string     `json:"provider_message_id" comment:"平台消息ID"`
	ErrorMessage      string     `json:"error_message" comment:"错误信息"`
	SentAt            *time.Time `json:"sent_at" comment:"发送时间"`
}
type WechatConfigItem struct {
	BaseItem `comment:"基础字段"`
	Name     string `json:"name" comment:"配置名称"`
	Type     int    `json:"type" comment:"微信应用类型"`
	AppID    string `json:"app_id" comment:"微信AppID"`
	Status   int    `json:"status" comment:"状态"`
	Remark   string `json:"remark" comment:"备注"`
}
type PaymentConfigItem struct {
	BaseItem     `comment:"基础字段"`
	Name         string `json:"name" comment:"配置名称"`
	Channel      int    `json:"channel" comment:"支付渠道"`
	AppID        string `json:"app_id" comment:"应用ID"`
	MerchantID   string `json:"merchant_id" comment:"商户号"`
	CertSerialNo string `json:"cert_serial_no" comment:"证书序列号"`
	NotifyURL    string `json:"notify_url" comment:"支付回调地址"`
	Status       int    `json:"status" comment:"状态"`
	Sort         int    `json:"sort" comment:"排序"`
	Remark       string `json:"remark" comment:"备注"`
}

type OperationLogItem struct {
	ID             uint      `json:"id" comment:"主键ID"`
	CreatedAt      time.Time `json:"created_at" comment:"操作时间"`
	UpdatedAt      time.Time `json:"updated_at" comment:"更新时间"`
	UserID         *uint     `json:"user_id" comment:"操作人ID"`
	Username       string    `json:"username" comment:"操作账号"`
	Method         string    `json:"method" comment:"请求方法"`
	Path           string    `json:"path" comment:"请求路由"`
	RequestParams  string    `json:"request_params" comment:"请求参数"`
	ResponseParams string    `json:"response_params" comment:"响应内容"`
	IP             string    `json:"ip" comment:"客户端IP"`
	UserAgent      string    `json:"user_agent" comment:"用户代理"`
	Code           int       `json:"code" comment:"业务响应码"`
	CostMs         int64     `json:"cost_ms" comment:"耗时毫秒"`
	Client         string    `json:"client" comment:"客户端类型"`
}
type LoginRecordItem struct {
	ID        uint       `json:"id" comment:"主键ID"`
	LoginID   string     `json:"login_id" comment:"登录标识"`
	UserID    uint       `json:"user_id" comment:"用户ID"`
	Username  string     `json:"username" comment:"账号"`
	Client    string     `json:"client" comment:"客户端类型"`
	LoginIP   string     `json:"login_ip" comment:"登录IP"`
	UserAgent string     `json:"user_agent" comment:"用户代理"`
	LoginAt   time.Time  `json:"login_at" comment:"登录时间"`
	LogoutAt  *time.Time `json:"logout_at" comment:"退出时间"`
	Status    int        `json:"status" comment:"登录状态"`
}

func base(item model.Base) BaseItem {
	return BaseItem{ID: item.ID, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, DeletedAt: item.DeletedAt}
}
func NewDeptItem(v model.SysDept) DeptItem {
	return DeptItem{BaseItem: base(v.Base), Name: v.Name, ParentID: v.ParentID, Sort: v.Sort, Leader: v.Leader, Remark: v.Remark}
}
func NewMenuItem(v model.SysMenu) MenuItem {
	return MenuItem{BaseItem: base(v.Base), Name: v.Name, Type: v.Type, ParentID: v.ParentID, Path: v.Path, ApiPath: v.ApiPath, Icon: v.Icon, Sort: v.Sort, Status: v.Status, Remark: v.Remark}
}
func NewRoleItem(v model.SysRole) RoleItem {
	return RoleItem{BaseItem: base(v.Base), Name: v.Name, Code: v.Code, ParentID: v.ParentID, Sort: v.Sort, Status: v.Status, Remark: v.Remark}
}
func NewUserItem(v model.SysUser, roles []model.SysRole) UserItem {
	roleItems := make([]RoleItem, 0, len(roles))
	for _, r := range roles {
		roleItems = append(roleItems, NewRoleItem(r))
	}
	return UserItem{BaseItem: base(v.Base), Username: v.Username, Nickname: v.Nickname, Avatar: v.Avatar, Mobile: v.Mobile, Email: v.Email, DeptID: v.DeptID, Status: v.Status, IsSuper: v.IsSuper, Roles: roleItems}
}
func NewStorageItem(v model.SysStorageConfig) StorageItem {
	return StorageItem{BaseItem: base(v.Base), Name: v.Name, Channel: v.Channel, Params: v.Params, IsDefault: v.IsDefault, Status: v.Status, Sort: v.Sort, Remark: v.Remark}
}
func NewSMSConfigItem(v model.SysSMSConfig) SMSConfigItem {
	return SMSConfigItem{BaseItem: base(v.Base), Name: v.Name, Provider: v.Provider, AccessKeyID: v.AccessKeyID, Endpoint: v.Endpoint, Status: v.Status, Remark: v.Remark}
}
func NewSMSSignatureItem(v model.SysSMSSignature) SMSSignatureItem {
	return SMSSignatureItem{BaseItem: base(v.Base), ConfigID: v.ConfigID, Name: v.Name, SignCode: v.SignCode, Status: v.Status, Remark: v.Remark}
}
func NewSMSTemplateItem(v model.SysSMSTemplate) SMSTemplateItem {
	return SMSTemplateItem{BaseItem: base(v.Base), ConfigID: v.ConfigID, Name: v.Name, TemplateCode: v.TemplateCode, Type: v.Type, Content: v.Content, Status: v.Status, Remark: v.Remark}
}
func NewSMSLogItem(v model.SysSMSSendLog) SMSLogItem {
	return SMSLogItem{ID: v.ID, CreatedAt: v.CreatedAt, ConfigID: v.ConfigID, SignatureID: v.SignatureID, TemplateID: v.TemplateID, Mobile: v.Mobile, Content: v.Content, Status: v.Status, ProviderMessageID: v.ProviderMessageID, ErrorMessage: v.ErrorMessage, SentAt: v.SentAt}
}
func NewWechatConfigItem(v model.SysWechatConfig) WechatConfigItem {
	return WechatConfigItem{BaseItem: base(v.Base), Name: v.Name, Type: v.Type, AppID: v.AppID, Status: v.Status, Remark: v.Remark}
}
func NewPaymentConfigItem(v model.SysPaymentConfig) PaymentConfigItem {
	return PaymentConfigItem{BaseItem: base(v.Base), Name: v.Name, Channel: v.Channel, AppID: v.AppID, MerchantID: v.MerchantID, CertSerialNo: v.CertSerialNo, NotifyURL: v.NotifyURL, Status: v.Status, Sort: v.Sort, Remark: v.Remark}
}
func NewOperationLogItem(v model.SysOperationLog) OperationLogItem {
	return OperationLogItem{ID: v.ID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, UserID: v.UserID, Username: v.Username, Method: v.Method, Path: v.Path, RequestParams: v.RequestParams, ResponseParams: v.ResponseParams, IP: v.IP, UserAgent: v.UserAgent, Code: v.Code, CostMs: v.CostMs, Client: v.Client}
}
func NewLoginRecordItem(v model.SysUserLogin) LoginRecordItem {
	return LoginRecordItem{ID: v.ID, LoginID: v.LoginID, UserID: v.UserID, Username: v.Username, Client: v.Client, LoginIP: v.LoginIP, UserAgent: v.UserAgent, LoginAt: v.LoginAt, LogoutAt: v.LogoutAt, Status: v.Status}
}

func NewDeptItems(list []model.SysDept) []DeptItem {
	result := make([]DeptItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewDeptItem(v))
	}
	return result
}
func NewMenuItems(list []model.SysMenu) []MenuItem {
	result := make([]MenuItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewMenuItem(v))
	}
	return result
}
func NewRoleItems(list []model.SysRole) []RoleItem {
	result := make([]RoleItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewRoleItem(v))
	}
	return result
}
func NewStorageItems(list []model.SysStorageConfig) []StorageItem {
	result := make([]StorageItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewStorageItem(v))
	}
	return result
}
func NewSMSConfigItems(list []model.SysSMSConfig) []SMSConfigItem {
	result := make([]SMSConfigItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewSMSConfigItem(v))
	}
	return result
}
func NewSMSSignatureItems(list []model.SysSMSSignature) []SMSSignatureItem {
	result := make([]SMSSignatureItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewSMSSignatureItem(v))
	}
	return result
}
func NewSMSTemplateItems(list []model.SysSMSTemplate) []SMSTemplateItem {
	result := make([]SMSTemplateItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewSMSTemplateItem(v))
	}
	return result
}
func NewSMSLogItems(list []model.SysSMSSendLog) []SMSLogItem {
	result := make([]SMSLogItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewSMSLogItem(v))
	}
	return result
}
func NewWechatConfigItems(list []model.SysWechatConfig) []WechatConfigItem {
	result := make([]WechatConfigItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewWechatConfigItem(v))
	}
	return result
}
func NewPaymentConfigItems(list []model.SysPaymentConfig) []PaymentConfigItem {
	result := make([]PaymentConfigItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewPaymentConfigItem(v))
	}
	return result
}
func NewOperationLogItems(list []model.SysOperationLog) []OperationLogItem {
	result := make([]OperationLogItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewOperationLogItem(v))
	}
	return result
}
func NewLoginRecordItems(list []model.SysUserLogin) []LoginRecordItem {
	result := make([]LoginRecordItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewLoginRecordItem(v))
	}
	return result
}
