package web

import (
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"unicode/utf8"
)

// UserHandler 我准备在它上面定义跟用户有关的路由
type UserHandler struct {
	svc         *service.UserService //组合service
	emailExp    *regexp.Regexp       //组合正则表达式
	passwordExp *regexp.Regexp
	birthdayExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
		birthDayRegexPattern = `^(19|20)\d{2}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	birthDayExp := regexp.MustCompile(birthDayRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		birthdayExp: birthDayExp,
	}
}

//func (u *UserHandler) RegisterRoutesV1(ug *gin.RouterGroup) {
//	ug.GET("/profile", u.Profile)
//	ug.GET("/profileJwt", u.ProfileJwt)
//	ug.POST("/signup", u.SignUp)
//	ug.POST("/login", u.loginJwt)
//	ug.POST("/edit", u.Edit)
//}

// RegisterRoutes 路由注册
func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.GET("/profile", u.ProfileJwt)
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.loginJwt)
	ug.POST("/edit", u.Edit)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	// Bind 方法会根据 Content-Type 来解析你的数据到 req 里面
	// 解析错了，就会直接写回一个 400 的错误
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "你的邮箱格式不对")
		return
	}
	if req.ConfirmPassword != req.Password {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		// 记录日志
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
		return
	}

	// 调用一下 svc 的方法
	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 步骤2
	// 在这里登录成功了
	// 设置 session
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 session 里面的值
	sess.Set("userId", user.Id)
	sess.Save()
	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) loginJwt(ctx *gin.Context) {
	type loginJwtReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req loginJwtReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//登录的时候直接返回jwt, 是返回在响应头中还是返回在响应体中?
	//判断是否登录成功
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "用户名或者密码错误")
		return
	}
	//生成token, 这边怎么生成过期时间?
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 2)),
		},
		UserId: user.Id,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) //传入自定义载荷
	mySigningKey := []byte("AllYourBase")
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误(jwt)")
		return
	}

	ctx.String(http.StatusOK, tokenString)

}

// Edit 编辑用户信息
func (u *UserHandler) Edit(ctx *gin.Context) {

	sess := sessions.Default(ctx)
	userId := sess.Get("userId")

	//定义请求的参数
	type EditReq struct {
		NickName string `json:"nickname"`
		Describe string `json:"describe"`
		BirthDay string `json:"birthday"`
	}

	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	//判断昵称的长度
	if utf8.RuneCountInString(req.NickName) > 20 {
		ctx.String(http.StatusOK, "昵称长度不能超过20个字符")
		return
	}

	if utf8.RuneCountInString(req.Describe) > 100 {
		ctx.String(http.StatusOK, "简介长度不能超过100个字符")
		return
	}

	ok, err := u.birthdayExp.MatchString(req.BirthDay) //使用正则去匹配
	if err != nil {
		// 记录日志, 可能是正则匹配超时了
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !ok {
		ctx.String(http.StatusOK, "生日不合法")
		return
	}

	//调用service方法去编辑用户
	err = u.svc.Edit(ctx, domain.User{
		NickName: req.NickName,
		BirthDay: req.BirthDay,
		Describe: req.Describe,
	}, userId.(int64))

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "编辑成功")

}

// Profile 获取用户信息
func (u *UserHandler) Profile(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	userId := sess.Get("userId")
	profile, err := u.svc.Profile(ctx, userId.(int64))
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":       profile.Id,
		"email":    profile.Email,
		"nickName": profile.NickName,
		"birthDay": profile.BirthDay,
		"describe": profile.Describe,
	})
}

func (u *UserHandler) ProfileJwt(ctx *gin.Context) {
	claims, ok := ctx.Get("claims") //从上下文中获取载荷,
	if !ok {
		//未登录
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	//sess := sessions.Default(ctx)
	//userId := sess.Get("userId")

	profile, err := u.svc.Profile(ctx, claims.(*UserClaims).UserId) //这边存的就是个指针,所以断言的时候这边也要是个指针
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":       profile.Id,
		"email":    profile.Email,
		"nickName": profile.NickName,
		"birthDay": profile.BirthDay,
		"describe": profile.Describe,
	})
}

// UserClaims 自定义载荷
type UserClaims struct {
	jwt.RegisteredClaims       //使用默认的一个实现类, 匿名组合
	UserId               int64 //自定义一个结构体去存放自定义的载荷
}
