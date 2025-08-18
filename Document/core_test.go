package document

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"refatorSiwu/domain"
)

// MockDocumentRepository Mock 文档仓储
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) Store(ctx context.Context, document *domain.Document) error {
	args := m.Called(ctx, document)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetByID(ctx context.Context, id int64) (*domain.Document, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Document), args.Error(1)
}

func (m *MockDocumentRepository) Update(ctx context.Context, document *domain.Document) error {
	args := m.Called(ctx, document)
	return args.Error(0)
}

func (m *MockDocumentRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDocumentRepository) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetByOwner(ctx context.Context, ownerID int64, includeDeleted bool) ([]*domain.Document, error) {
	args := m.Called(ctx, ownerID, includeDeleted)
	return args.Get(0).([]*domain.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetByParent(ctx context.Context, parentID *int64, ownerID int64) ([]*domain.Document, error) {
	args := m.Called(ctx, parentID, ownerID)
	return args.Get(0).([]*domain.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetBySpace(ctx context.Context, spaceID int64, ownerID int64) ([]*domain.Document, error) {
	args := m.Called(ctx, spaceID, ownerID)
	return args.Get(0).([]*domain.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetDocumentTree(ctx context.Context, rootID *int64, ownerID int64) ([]*domain.Document, error) {
	args := m.Called(ctx, rootID, ownerID)
	return args.Get(0).([]*domain.Document), args.Error(1)
}

func (m *MockDocumentRepository) SearchDocuments(ctx context.Context, userID int64, keyword string, docType *domain.DocumentType, limit, offset int) ([]*domain.Document, error) {
	args := m.Called(ctx, userID, keyword, docType, limit, offset)
	return args.Get(0).([]*domain.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetStarredDocuments(ctx context.Context, userID int64) ([]*domain.Document, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetRecentDocuments(ctx context.Context, userID int64, limit int) ([]*domain.Document, error) {
	args := m.Called(ctx, userID, limit)
	return args.Get(0).([]*domain.Document), args.Error(1)
}

func (m *MockDocumentRepository) UpdateContent(ctx context.Context, id int64, content string) error {
	args := m.Called(ctx, id, content)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetContent(ctx context.Context, id int64) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *MockDocumentRepository) UpdateStatus(ctx context.Context, id int64, status domain.DocumentStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockDocumentRepository) ToggleStar(ctx context.Context, id int64, userID int64, starred bool) error {
	args := m.Called(ctx, id, userID, starred)
	return args.Error(0)
}

func (m *MockDocumentRepository) MoveDocument(ctx context.Context, id int64, newParentID *int64) error {
	args := m.Called(ctx, id, newParentID)
	return args.Error(0)
}

func (m *MockDocumentRepository) BatchDelete(ctx context.Context, ids []int64, userID int64) error {
	args := m.Called(ctx, ids, userID)
	return args.Error(0)
}

func (m *MockDocumentRepository) BatchMove(ctx context.Context, ids []int64, newParentID *int64, userID int64) error {
	args := m.Called(ctx, ids, newParentID, userID)
	return args.Error(0)
}

// MockUserRepository Mock 用户仓储
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Store(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByGitHubID(ctx context.Context, githubID int64) (*domain.User, error) {
	args := m.Called(ctx, githubID)
	return args.Get(0).(*domain.User), args.Error(1)
}

// Mock 子域服务
type MockDocumentShareUsecase struct {
	mock.Mock
}

func (m *MockDocumentShareUsecase) ShareDocument(ctx context.Context, userID, documentID int64, permission domain.Permission, password string, expiresAt *time.Time, shareWithUserIDs []int64) (*domain.DocumentShare, error) {
	args := m.Called(ctx, userID, documentID, permission, password, expiresAt, shareWithUserIDs)
	return args.Get(0).(*domain.DocumentShare), args.Error(1)
}

func (m *MockDocumentShareUsecase) UpdateShareLink(ctx context.Context, userID, shareID int64, permission *domain.Permission, password *string, expiresAt *time.Time) (*domain.DocumentShare, error) {
	args := m.Called(ctx, userID, shareID, permission, password, expiresAt)
	return args.Get(0).(*domain.DocumentShare), args.Error(1)
}

func (m *MockDocumentShareUsecase) DeleteShareLink(ctx context.Context, userID, shareID int64) error {
	args := m.Called(ctx, userID, shareID)
	return args.Error(0)
}

func (m *MockDocumentShareUsecase) GetDocumentShares(ctx context.Context, userID, documentID int64) ([]*domain.DocumentShare, error) {
	args := m.Called(ctx, userID, documentID)
	return args.Get(0).([]*domain.DocumentShare), args.Error(1)
}

func (m *MockDocumentShareUsecase) GetMySharedDocuments(ctx context.Context, userID int64) ([]*domain.DocumentShare, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.DocumentShare), args.Error(1)
}

func (m *MockDocumentShareUsecase) GetSharedDocument(ctx context.Context, linkID, password string, accessIP string) (*domain.DocumentAccessInfo, error) {
	args := m.Called(ctx, linkID, password, accessIP)
	return args.Get(0).(*domain.DocumentAccessInfo), args.Error(1)
}

func (m *MockDocumentShareUsecase) ValidateShareAccess(ctx context.Context, linkID, password string) (*domain.DocumentShare, error) {
	args := m.Called(ctx, linkID, password)
	return args.Get(0).(*domain.DocumentShare), args.Error(1)
}

type MockDocumentPermissionUsecase struct {
	mock.Mock
}

func (m *MockDocumentPermissionUsecase) GrantDocumentPermission(ctx context.Context, userID, documentID, targetUserID int64, permission domain.Permission) error {
	args := m.Called(ctx, userID, documentID, targetUserID, permission)
	return args.Error(0)
}

func (m *MockDocumentPermissionUsecase) RevokeDocumentPermission(ctx context.Context, userID, documentID, targetUserID int64) error {
	args := m.Called(ctx, userID, documentID, targetUserID)
	return args.Error(0)
}

func (m *MockDocumentPermissionUsecase) UpdateDocumentPermission(ctx context.Context, userID, documentID, targetUserID int64, permission domain.Permission) error {
	args := m.Called(ctx, userID, documentID, targetUserID, permission)
	return args.Error(0)
}

func (m *MockDocumentPermissionUsecase) CheckDocumentPermission(ctx context.Context, userID, documentID int64, permission domain.Permission) (bool, error) {
	args := m.Called(ctx, userID, documentID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockDocumentPermissionUsecase) GetDocumentPermissions(ctx context.Context, userID, documentID int64) ([]*domain.DocumentPermission, error) {
	args := m.Called(ctx, userID, documentID)
	return args.Get(0).([]*domain.DocumentPermission), args.Error(1)
}

func (m *MockDocumentPermissionUsecase) BatchGrantPermission(ctx context.Context, userID, documentID int64, targetUserIDs []int64, permission domain.Permission) error {
	args := m.Called(ctx, userID, documentID, targetUserIDs, permission)
	return args.Error(0)
}

func (m *MockDocumentPermissionUsecase) BatchRevokePermission(ctx context.Context, userID, documentID int64, targetUserIDs []int64) error {
	args := m.Called(ctx, userID, documentID, targetUserIDs)
	return args.Error(0)
}

type MockDocumentFavoriteUsecase struct {
	mock.Mock
}

func (m *MockDocumentFavoriteUsecase) ToggleDocumentFavorite(ctx context.Context, userID, documentID int64) (bool, error) {
	args := m.Called(ctx, userID, documentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDocumentFavoriteUsecase) SetFavoriteCustomTitle(ctx context.Context, userID, documentID int64, customTitle string) error {
	args := m.Called(ctx, userID, documentID, customTitle)
	return args.Error(0)
}

func (m *MockDocumentFavoriteUsecase) RemoveDocumentFavorite(ctx context.Context, userID, documentID int64) error {
	args := m.Called(ctx, userID, documentID)
	return args.Error(0)
}

func (m *MockDocumentFavoriteUsecase) GetFavoriteDocuments(ctx context.Context, userID int64) ([]*domain.DocumentFavorite, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.DocumentFavorite), args.Error(1)
}

func (m *MockDocumentFavoriteUsecase) IsFavoriteDocument(ctx context.Context, userID, documentID int64) (bool, error) {
	args := m.Called(ctx, userID, documentID)
	return args.Bool(0), args.Error(1)
}

// 测试用例

func TestCreateDocument_Success(t *testing.T) {
	// 准备 Mock
	mockDocRepo := new(MockDocumentRepository)
	mockUserRepo := new(MockUserRepository)
	mockShareUsecase := new(MockDocumentShareUsecase)
	mockPermUsecase := new(MockDocumentPermissionUsecase)
	mockFavoriteUsecase := new(MockDocumentFavoriteUsecase)

	// 创建服务实例
	service := NewDocumentService(
		mockDocRepo,
		mockShareUsecase,
		mockPermUsecase,
		mockFavoriteUsecase,
		mockUserRepo,
	)

	// 准备测试数据
	ctx := context.Background()
	userID := int64(1)
	user := &domain.User{ID: userID, Username: "testuser"}

	// 设置 Mock 期望
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockDocRepo.On("Store", ctx, mock.AnythingOfType("*domain.Document")).Return(nil)

	// 执行测试
	document, err := service.CreateDocument(
		ctx,
		userID,
		"测试文档",
		"{}",
		domain.DocumentTypeFile,
		nil,
		nil,
		0,
		false,
	)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, document)
	assert.Equal(t, "测试文档", document.Title)
	assert.Equal(t, domain.DocumentTypeFile, document.Type)
	assert.Equal(t, userID, document.OwnerID)

	// 验证 Mock 调用
	mockUserRepo.AssertExpectations(t)
	mockDocRepo.AssertExpectations(t)
}

func TestCreateDocument_UserNotFound(t *testing.T) {
	// 准备 Mock
	mockDocRepo := new(MockDocumentRepository)
	mockUserRepo := new(MockUserRepository)
	mockShareUsecase := new(MockDocumentShareUsecase)
	mockPermUsecase := new(MockDocumentPermissionUsecase)
	mockFavoriteUsecase := new(MockDocumentFavoriteUsecase)

	// 创建服务实例
	service := NewDocumentService(
		mockDocRepo,
		mockShareUsecase,
		mockPermUsecase,
		mockFavoriteUsecase,
		mockUserRepo,
	)

	// 准备测试数据
	ctx := context.Background()
	userID := int64(999)

	// 设置 Mock 期望
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, domain.ErrUserNotFound)

	// 执行测试
	document, err := service.CreateDocument(
		ctx,
		userID,
		"测试文档",
		"{}",
		domain.DocumentTypeFile,
		nil,
		nil,
		0,
		false,
	)

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserNotFound, err)
	assert.Nil(t, document)

	// 验证 Mock 调用
	mockUserRepo.AssertExpectations(t)
}

func TestCreateDocument_InvalidTitle(t *testing.T) {
	// 准备 Mock
	mockDocRepo := new(MockDocumentRepository)
	mockUserRepo := new(MockUserRepository)
	mockShareUsecase := new(MockDocumentShareUsecase)
	mockPermUsecase := new(MockDocumentPermissionUsecase)
	mockFavoriteUsecase := new(MockDocumentFavoriteUsecase)

	// 创建服务实例
	service := NewDocumentService(
		mockDocRepo,
		mockShareUsecase,
		mockPermUsecase,
		mockFavoriteUsecase,
		mockUserRepo,
	)

	// 准备测试数据
	ctx := context.Background()
	userID := int64(1)
	user := &domain.User{ID: userID, Username: "testuser"}

	// 设置 Mock 期望
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// 执行测试（空标题）
	document, err := service.CreateDocument(
		ctx,
		userID,
		"", // 空标题
		"{}",
		domain.DocumentTypeFile,
		nil,
		nil,
		0,
		false,
	)

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidDocumentTitle, err)
	assert.Nil(t, document)

	// 验证 Mock 调用
	mockUserRepo.AssertExpectations(t)
}

func TestGetDocument_Success(t *testing.T) {
	// 准备 Mock
	mockDocRepo := new(MockDocumentRepository)
	mockUserRepo := new(MockUserRepository)
	mockShareUsecase := new(MockDocumentShareUsecase)
	mockPermUsecase := new(MockDocumentPermissionUsecase)
	mockFavoriteUsecase := new(MockDocumentFavoriteUsecase)

	// 创建服务实例
	service := NewDocumentService(
		mockDocRepo,
		mockShareUsecase,
		mockPermUsecase,
		mockFavoriteUsecase,
		mockUserRepo,
	)

	// 准备测试数据
	ctx := context.Background()
	userID := int64(1)
	documentID := int64(100)
	document := &domain.Document{
		ID:      documentID,
		Title:   "测试文档",
		OwnerID: userID,
		Status:  domain.DocumentStatusActive,
	}

	// 设置 Mock 期望
	mockDocRepo.On("GetByID", ctx, documentID).Return(document, nil)
	mockPermUsecase.On("CheckDocumentPermission", ctx, userID, documentID, domain.PermissionView).Return(true, nil)

	// 执行测试
	result, err := service.GetDocument(ctx, userID, documentID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, documentID, result.ID)
	assert.Equal(t, "测试文档", result.Title)

	// 验证 Mock 调用
	mockDocRepo.AssertExpectations(t)
	mockPermUsecase.AssertExpectations(t)
}

func TestGetDocument_NotFound(t *testing.T) {
	// 准备 Mock
	mockDocRepo := new(MockDocumentRepository)
	mockUserRepo := new(MockUserRepository)
	mockShareUsecase := new(MockDocumentShareUsecase)
	mockPermUsecase := new(MockDocumentPermissionUsecase)
	mockFavoriteUsecase := new(MockDocumentFavoriteUsecase)

	// 创建服务实例
	service := NewDocumentService(
		mockDocRepo,
		mockShareUsecase,
		mockPermUsecase,
		mockFavoriteUsecase,
		mockUserRepo,
	)

	// 准备测试数据
	ctx := context.Background()
	userID := int64(1)
	documentID := int64(999)

	// 设置 Mock 期望
	mockDocRepo.On("GetByID", ctx, documentID).Return(nil, domain.ErrDocumentNotFound)

	// 执行测试
	result, err := service.GetDocument(ctx, userID, documentID)

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, domain.ErrDocumentNotFound, err)
	assert.Nil(t, result)

	// 验证 Mock 调用
	mockDocRepo.AssertExpectations(t)
}

// 运行测试：go test ./document -v
