package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"DOC/config"
	"DOC/domain"
	"DOC/pkg/jwt"
	"DOC/pkg/utils"
)

// authService 认证业务服务实现
// 专注于用户认证、会话管理和令牌处理
type authService struct {
	userRepo       domain.UserRepository      // 用户仓储接口
	authRepo       domain.AuthRepository      // 认证仓储接口
	authCache      domain.AuthCacheRepository // 认证缓存仓储接口
	emailService   domain.EmailUsecase        // 邮件服务接口
	jwtManager     *jwt.JWTManager            // JWT 管理器
	config         *config.Config             // 应用配置
	contextTimeout time.Duration              // 上下文超时时间
}

// NewAuthService 创建新的认证服务实例
func NewAuthService(
	userRepo domain.UserRepository,
	authRepo domain.AuthRepository,
	authCache domain.AuthCacheRepository,
	emailService domain.EmailUsecase,
	jwtManager *jwt.JWTManager,
	config *config.Config,
	timeout time.Duration,
) domain.AuthUsecase {
	return &authService{
		userRepo:       userRepo,
		authRepo:       authRepo,
		authCache:      authCache,
		emailService:   emailService,
		jwtManager:     jwtManager,
		config:         config,
		contextTimeout: timeout,
	}
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, user *domain.User) (*domain.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 验证用户实体
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// 2. 检查邮箱唯一性
	if existUser, err := s.userRepo.GetByEmail(ctx, user.Email); err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			fmt.Printf("检查邮箱唯一性失败: %v", err)
			return nil, domain.ErrInternalError
		}
	} else if existUser != nil {
		return nil, domain.ErrUserAlreadyExist
	}

	// 3. 检查用户名唯一性
	if existUser, err := s.userRepo.GetByUsername(ctx, user.Username); err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			fmt.Printf("检查用户名唯一性失败: %v", err)
			return nil, domain.ErrInternalError
		}
	} else if existUser != nil {
		return nil, domain.ErrUserAlreadyExist
	}

	// 4. 设置默认值
	user.Status = domain.UserStatusActive
	if user.Role == "" {
		user.Role = "user"
	}

	// 5. 保存用户到数据库
	if err := s.userRepo.Store(ctx, user); err != nil {
		fmt.Printf("保存用户失败: %v", err)
		return nil, domain.ErrInternalError
	}

	// 6. 生成令牌和创建会话
	return s.createAuthSession(ctx, user, domain.AuthProviderEmail)
}

// Login 密码登录
func (s *authService) Login(ctx context.Context, email, password string) (*domain.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 参数验证
	if email == "" || password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	// 2. 查找用户
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		fmt.Printf("查找用户失败: %v", err)
		return nil, domain.ErrInternalError
	}

	// 3. 检查用户状态
	if !user.CanLogin() {
		return nil, domain.ErrUserNotActive
	}

	// 4. 验证密码
	if !utils.CheckPassword(password, user.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	// 5. 更新最后登录时间
	user.UpdateLastLogin()
	if err := s.userRepo.Update(ctx, user); err != nil {
		// 记录日志但不阻断登录流程
		fmt.Printf("更新最后登录时间失败: %v\n", err)
	}

	// 6. 生成令牌和创建会话
	return s.createAuthSession(ctx, user, domain.AuthProviderEmail)
}

// VerifyCode 验证码登录/注册
func (s *authService) VerifyCode(ctx context.Context, email, code string) (*domain.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 参数验证
	if email == "" || code == "" {
		return nil, domain.ErrInvalidVerificationCode
	}

	// 2. 验证邮箱格式
	if !utils.IsValidEmail(email) {
		return nil, domain.ErrInvalidUserEmail
	}

	// 3. 验证验证码
	storedCode, err := s.authCache.GetEmailCode(ctx, email)
	if err != nil {
		if !errors.Is(err, domain.ErrVerificationCodeNotFound) {
			return nil, domain.ErrInternalError
		}
		// 日志打印
		fmt.Printf("%v", err)
		return nil, domain.ErrVerificationCodeNotFound
	}
	if storedCode != code {
		return nil, domain.ErrInvalidVerificationCode
	}

	// 4. 删除已使用的验证码
	if err := s.authCache.DeleteEmailCode(ctx, email); err != nil {
		// 日志打印
		fmt.Printf("删除验证码失败: %v\n", err)
	}

	// 5. 查找或创建用户
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// 用户不存在，创建新用户（自动注册）
			user = &domain.User{
				Email:    email,
				Username: email, // 使用邮箱作为用户名
				Name:     email,
				Role:     "user",
				Status:   domain.UserStatusActive,
			}

			if err := s.userRepo.Store(ctx, user); err != nil {
				fmt.Printf("%v\n", err)
				return nil, domain.ErrInternalError
			}
		} else {
			fmt.Printf("%v\n", err)
			return nil, domain.ErrInternalError
		}
	}

	// 6. 检查用户状态
	if !user.CanLogin() {
		return nil, domain.ErrUserNotActive
	}

	// 7. 更新最后登录时间
	user.UpdateLastLogin()
	if err := s.userRepo.Update(ctx, user); err != nil {
		fmt.Printf("更新最后登录时间失败: %v\n", err)
	}

	// 8. 生成令牌和创建会话
	return s.createAuthSession(ctx, user, domain.AuthProviderCode)
}

// SendVerificationCode 发送验证码
func (s *authService) SendVerificationCode(ctx context.Context, email string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 验证邮箱格式
	if !utils.IsValidEmail(email) {
		return domain.ErrInvalidUserEmail
	}

	// 2. 检查发送间隔
	if canSend, err := s.authCache.CheckSendInterval(ctx, email); err != nil {
		fmt.Printf("检查发送间隔失败: %v", err)
		return domain.ErrInternalError
	} else if !canSend {
		return domain.ErrVerificationCodeSendTooFrequent
	}

	// 3. 生成验证码
	code, err := utils.GenerateRandomCode(6)
	if err != nil {
		fmt.Printf("生成验证码失败: %w", err)
		return domain.ErrInternalError
	}

	// 4. 存储验证码到缓存
	if err := s.authCache.StoreEmailCode(ctx, email, code, 10*time.Minute); err != nil {
		fmt.Printf("存储验证码失败: %w", err)
		return domain.ErrInternalError
	}

	// 5. 设置发送间隔
	if err := s.authCache.SetSendInterval(ctx, email, 1*time.Minute); err != nil {
		fmt.Printf("设置发送间隔失败: %v\n", err)
	}

	// 6. 发送邮件
	if err := s.emailService.SendVerificationEmail(ctx, email, code); err != nil {
		fmt.Printf("发送邮件失败: %v", err)
		return domain.ErrInternalError
	}

	return nil
}

// createAuthSession 创建认证会话并生成令牌
func (s *authService) createAuthSession(ctx context.Context, user *domain.User, provider domain.AuthProvider) (*domain.AuthResponse, error) {
	// 1. 生成令牌
	accessToken, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		fmt.Printf("生成令牌失败: %v", err)
		return nil, domain.ErrInternalError
	}

	// 2. 生成刷新令牌（简化处理，使用相同方法）
	refreshToken, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		fmt.Printf("生成刷新令牌失败: %v", err)
		return nil, domain.ErrInternalError
	}

	// 3. 创建会话记录
	session := &domain.AuthSession{
		UserID:       user.ID,
		SessionToken: accessToken,
		RefreshToken: refreshToken,
		Provider:     provider,
		Status:       domain.SessionStatusActive,
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24小时过期
	}

	// 4. 保存会话
	if err := s.authRepo.StoreSession(ctx, session); err != nil {
		fmt.Printf("保存会话失败%v\n", err)
		return nil, domain.ErrInternalError
	}

	// 5. 返回认证响应
	return &domain.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    24 * 60 * 60 * 1000, // 24小时，毫秒
	}, nil
}

// ValidateToken 验证访问令牌
func (s *authService) ValidateToken(ctx context.Context, accessToken string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 验证令牌格式和签名
	claims, err := s.jwtManager.ValidateToken(accessToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// 2. 检查会话是否存在且有效
	session, err := s.authRepo.GetSessionByToken(ctx, accessToken)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if !session.IsActive() {
		return nil, domain.ErrSessionExpired
	}

	// 3. 获取用户信息
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v", err)
		return nil, domain.ErrInternalError
	}

	// 4. 检查用户状态
	if !user.CanLogin() {
		return nil, domain.ErrUserNotActive
	}

	// 5. 更新会话最后使用时间
	session.UpdateLastUsed()
	if err := s.authRepo.UpdateSession(ctx, session); err != nil {
		fmt.Printf("更新会话使用时间失败: %v\n", err)
	}

	return user, nil
}

// RefreshToken 刷新访问令牌
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 1. 验证刷新令牌
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// 2. 查找会话
	sessions, err := s.authRepo.GetSessionsByUserID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	var session *domain.AuthSession
	for _, s := range sessions {
		if s.RefreshToken == refreshToken && s.IsActive() {
			session = s
			break
		}
	}

	if session == nil {
		return nil, domain.ErrSessionExpired
	}

	// 3. 获取用户信息
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v", err)
		return nil, domain.ErrInternalError
	}

	if !user.CanLogin() {
		return nil, domain.ErrUserNotActive
	}

	// 4. 生成新的访问令牌
	accessToken, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		fmt.Printf("生成令牌失败: %v", err)
		return nil, domain.ErrInternalError
	}

	// 5. 更新会话
	session.SessionToken = accessToken
	session.UpdateLastUsed()
	if err := s.authRepo.UpdateSession(ctx, session); err != nil {
		fmt.Printf("更新会话失败: %v", err)
		return nil, domain.ErrInternalError
	}

	return &domain.TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   24 * 60 * 60 * 1000, // 24小时，毫秒
	}, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, sessionToken string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 查找并撤销会话
	session, err := s.authRepo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		return domain.ErrSessionNotFound
	}
	if err := s.authRepo.RevokeSession(ctx, session.ID); err != nil {
		if !errors.Is(err, domain.ErrSessionNotFound) {
			fmt.Printf("%v", err)
			return domain.ErrInternalError
		}
		return domain.ErrSessionNotFound
	}
	return nil
}

// LogoutAll 用户全部登出
func (s *authService) LogoutAll(ctx context.Context, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if err := s.authRepo.RevokeAllUserSessions(ctx, userID); err != nil {
		fmt.Printf("%v\n", err)
		return domain.ErrInternalError
	}
	return nil
}

// GenerateOAuthURL 生成OAuth授权URL
func (s *authService) GenerateOAuthURL(ctx context.Context, provider domain.AuthProvider) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 检查支持的提供商
	if provider != domain.AuthProviderGitHub {
		fmt.Printf("不支持的OAuth提供商: %v", provider)
		return "", "", domain.ErrOauthNotSupported
	}

	// 获取GitHub OAuth配置
	githubCfg := s.config.OAuth.GitHub
	if githubCfg.ClientID == "" || githubCfg.ClientSecret == "" {
		fmt.Printf("GitHub OAuth配置不完整")
		return "", "", domain.ErrOAuthUserInfoFailed
	}

	// 生成随机 state 参数，用于防止CSRF攻击
	state, err := utils.GenerateRandomString(32)
	if err != nil {
		fmt.Printf("生成state失败: %v", err)
		return "", "", domain.ErrOAuthUserInfoFailed
	}

	// 构建授权URL参数
	params := url.Values{
		"client_id":     {githubCfg.ClientID},
		"redirect_uri":  {githubCfg.APICallbackURL}, // 使用后端回调URL
		"scope":         {strings.Join(githubCfg.Scopes, " ")},
		"state":         {state},
		"response_type": {"code"},
	}

	// 构建完整的授权URL
	authURL := fmt.Sprintf("%s?%s", githubCfg.AuthURL, params.Encode())

	// 存储OAuth state到数据库，用于后续验证
	oauthState := &domain.OAuthState{
		State:     state,
		Provider:  provider,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10分钟过期
	}

	if err := s.authRepo.StoreOAuthState(ctx, oauthState); err != nil {
		fmt.Printf("存储OAuth state失败: %v", err)
		return "", "", domain.ErrOAuthUserInfoFailed
	}

	return authURL, state, nil
}

// HandleOAuthCallback 处理OAuth回调
func (s *authService) HandleOAuthCallback(ctx context.Context, provider domain.AuthProvider, code, state string) (*domain.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// 检查支持的提供商
	if provider != domain.AuthProviderGitHub {
		fmt.Printf("不支持的OAuth提供商: %v", provider)
		return nil, domain.ErrOauthNotSupported
	}

	// 获取GitHub OAuth配置
	githubCfg := s.config.OAuth.GitHub
	// 验证state参数，防止CSRF攻击
	storedState, err := s.authRepo.GetOAuthState(ctx, state)
	if err != nil {
		return nil, domain.ErrOAuthStateMismatch
	}
	if storedState.State != state {
		return nil, domain.ErrOAuthStateMismatch
	}

	// 删除已使用的state
	if err := s.authRepo.DeleteOAuthState(ctx, storedState.ID); err != nil {
		// 记录日志但不阻断流程
		fmt.Printf("删除OAuth state失败: %v\n", err)
	}

	// 步骤1: 交换访问令牌
	accessToken, err := s.exchangeCodeForToken(githubCfg, code)
	if err != nil {
		fmt.Printf("交换访问令牌失败: %v", err)
		return nil, domain.ErrInternalError
	}

	// 步骤2: 获取用户信息
	userInfo, err := s.fetchGitHubUserInfo(githubCfg, accessToken)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v", err)
		return nil, domain.ErrInternalError
	}

	// 步骤3: 查找或创建用户
	user, err := s.findOrCreateOAuthUser(ctx, userInfo, provider)
	if err != nil {
		fmt.Printf("处理用户失败: %v", err)
		return nil, domain.ErrInternalError
	}

	// 步骤4: 检查用户状态
	if !user.CanLogin() {
		return nil, domain.ErrUserNotActive
	}

	// 步骤5: 更新最后登录时间
	user.UpdateLastLogin()
	if err := s.userRepo.Update(ctx, user); err != nil {
		fmt.Printf("更新最后登录时间失败: %v\n", err)
	}

	// 步骤6: 创建认证会话
	return s.createAuthSession(ctx, user, domain.AuthProviderGitHub)
}

// GetFrontedURL 返回前端回调URL
func (s *authService) GetFrontedURL(response *domain.AuthResponse) string {
	fmt.Println(s.config.OAuth.GitHub.RedirectURL + "?token=" + response.AccessToken)
	return s.config.OAuth.GitHub.RedirectURL + "?token=" + response.AccessToken
}

// RevokeToken 撤销令牌
func (s *authService) RevokeToken(ctx context.Context, sessionToken string) error {
	return s.Logout(ctx, sessionToken)
}

// GetUserSessions 获取用户会话列表
func (s *authService) GetUserSessions(ctx context.Context, userID int64) ([]*domain.AuthSession, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	authSession, err := s.authRepo.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, domain.ErrInternalError
	}
	return authSession, nil
}

// RevokeSession 撤销特定会话
func (s *authService) RevokeSession(ctx context.Context, userID int64, sessionID int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if err := s.authRepo.RevokeSession(ctx, sessionID); err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return domain.ErrSessionNotFound
		}
		return domain.ErrInternalError
	}

	return nil

}

// === OAuth 辅助方法 ===

// exchangeCodeForToken 使用授权码交换访问令牌
func (s *authService) exchangeCodeForToken(githubCfg config.GitHubOAuthConfig, code string) (string, error) {
	// 构建token交换请求
	params := url.Values{
		"client_id":     {githubCfg.ClientID},
		"client_secret": {githubCfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {githubCfg.APICallbackURL},
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Duration(githubCfg.Timeout) * time.Second,
	}

	// 发送POST请求
	resp, err := client.PostForm(githubCfg.TokenURL, params)
	if err != nil {
		return "", fmt.Errorf("发送token交换请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取token响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token交换失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应（GitHub返回URL编码格式）
	tokenParams, err := url.ParseQuery(string(body))
	if err != nil {
		return "", fmt.Errorf("解析token响应失败: %w", err)
	}

	accessToken := tokenParams.Get("access_token")
	if accessToken == "" {
		errorDesc := tokenParams.Get("error_description")
		if errorDesc == "" {
			errorDesc = "未知错误"
		}
		return "", fmt.Errorf("无法获取访问令牌: %s", errorDesc)
	}

	return accessToken, nil
}

// fetchGitHubUserInfo 获取GitHub用户信息
func (s *authService) fetchGitHubUserInfo(githubCfg config.GitHubOAuthConfig, accessToken string) (*domain.OAuthUserInfo, error) {
	// 创建获取用户信息的请求
	userInfoURL := githubCfg.APIBaseURL + "/user"
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		fmt.Printf("创建用户信息请求失败: %v", err)
		return nil, err
	}

	// 设置授权头
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/json")

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(githubCfg.Timeout) * time.Second,
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("获取用户信息失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
		return nil, err
	}

	// 解析用户信息
	var githubUser struct {
		ID       int    `json:"id"`
		Login    string `json:"login"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar_url"`
		HTMLURL  string `json:"html_url"`
		Company  string `json:"company"`
		Location string `json:"location"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		fmt.Printf("解析用户信息失败: %v", err)
		return nil, err
	}

	// 如果主要邮箱为空，尝试获取邮箱列表
	if githubUser.Email == "" {
		email, err := s.fetchGitHubUserEmail(githubCfg, accessToken)
		if err != nil {
			fmt.Printf("获取用户邮箱失败: %v", err)
			return nil, err

		}
		githubUser.Email = email
	}

	// 转换为领域对象
	userInfo := &domain.OAuthUserInfo{
		ID:       fmt.Sprintf("%d", githubUser.ID),
		Username: githubUser.Login,
		Name:     githubUser.Name,
		Email:    githubUser.Email,
		Avatar:   githubUser.Avatar,
	}

	// 如果姓名为空，使用用户名
	if userInfo.Name == "" {
		userInfo.Name = userInfo.Username
	}

	return userInfo, nil
}

// fetchGitHubUserEmail 获取GitHub用户的邮箱地址
func (s *authService) fetchGitHubUserEmail(githubCfg config.GitHubOAuthConfig, accessToken string) (string, error) {
	emailURL := githubCfg.APIBaseURL + "/user/emails"
	req, err := http.NewRequest("GET", emailURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: time.Duration(githubCfg.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("获取邮箱列表失败，状态码: %v", resp.StatusCode)
		return "", err
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// 查找主要邮箱
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// 如果没有主要邮箱，返回第一个已验证的邮箱
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	fmt.Printf("未找到已验证的邮箱地址")
	return "", err
}

// findOrCreateOAuthUser 查找或创建OAuth用户
func (s *authService) findOrCreateOAuthUser(ctx context.Context, userInfo *domain.OAuthUserInfo, provider domain.AuthProvider) (*domain.User, error) {
	// 首先尝试通过邮箱查找用户
	user, err := s.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// 用户不存在，创建新用户
			user = &domain.User{
				Username:  userInfo.Username,
				Email:     userInfo.Email,
				Name:      userInfo.Name,
				AvatarURL: userInfo.Avatar,
				GitHubID:  userInfo.ID,
				Role:      "user",
				Status:    domain.UserStatusActive,
			}

			// 确保用户名唯一性
			if existUser, _ := s.userRepo.GetByUsername(ctx, user.Username); existUser != nil {
				// 用户名已存在，添加后缀
				user.Username = fmt.Sprintf("%s_%s", user.Username, userInfo.ID)
			}

			// 保存新用户
			if err := s.userRepo.Store(ctx, user); err != nil {
				fmt.Printf("创建OAuth用户失败: %v", err)
				return nil, err
			}

			return user, nil
		}
		fmt.Printf("查找用户失败: %v", err)
		return nil, err
	}

	// 用户已存在，更新OAuth相关信息
	if user.GitHubID == "" {
		user.GitHubID = userInfo.ID
	}

	// 更新头像信息
	if userInfo.Avatar != "" {
		user.AvatarURL = userInfo.Avatar
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		fmt.Printf("更新OAuth用户信息失败: %v\n", err)
	}

	return user, nil
}
