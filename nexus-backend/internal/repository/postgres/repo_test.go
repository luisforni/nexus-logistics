package postgres

import (
"context"
"database/sql"
"fmt"
"regexp"
"testing"
"time"

"github.com/DATA-DOG/go-sqlmock"
"github.com/google/uuid"
"github.com/luisforni/nexus-logistics/backend/internal/domain"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"gorm.io/driver/postgres"
"gorm.io/gorm"
"gorm.io/gorm/logger"
)

func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
t.Helper()
sqlDB, mock, err := sqlmock.New()
require.NoError(t, err)
gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
Logger: logger.Default.LogMode(logger.Silent),
})
require.NoError(t, err)
t.Cleanup(func() { sqlDB.Close() })
return gormDB, mock
}

func TestShipmentRepo_New(t *testing.T) {
db, _ := newMockDB(t)
assert.NotNil(t, NewShipmentRepository(db))
}

func TestShipmentRepo_Create_Success(t *testing.T) {
db, mock := newMockDB(t)
repo := NewShipmentRepository(db)
s := &domain.Shipment{ID: uuid.New(), SenderID: uuid.New(), TrackingNumber: "NX-x", Status: domain.StatusPending, RecipientName: "A", EstimatedAt: time.Now()}
mock.ExpectBegin()
mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shipments"`)).
WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(time.Now(), time.Now()))
mock.ExpectCommit()
assert.NoError(t, repo.Create(context.Background(), s))
}

func TestShipmentRepo_Create_Error(t *testing.T) {
db, mock := newMockDB(t)
repo := NewShipmentRepository(db)
mock.ExpectBegin()
mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shipments"`)).WillReturnError(sql.ErrConnDone)
mock.ExpectRollback()
assert.Error(t, repo.Create(context.Background(), &domain.Shipment{ID: uuid.New(), SenderID: uuid.New()}))
}

func TestShipmentRepo_FindByID_NotFound(t *testing.T) {
db, mock := newMockDB(t)
repo := NewShipmentRepository(db)
id := uuid.New()
mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WithArgs(id, 1).WillReturnRows(sqlmock.NewRows(nil))
_, err := repo.FindByID(context.Background(), id)
require.Error(t, err)
assert.Contains(t, err.Error(), "not found")
}

func TestShipmentRepo_FindByTrackingNumber_NotFound(t *testing.T) {
db, mock := newMockDB(t)
repo := NewShipmentRepository(db)
mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WithArgs("NX-unk", 1).WillReturnRows(sqlmock.NewRows(nil))
_, err := repo.FindByTrackingNumber(context.Background(), "NX-unk")
require.Error(t, err)
assert.Contains(t, err.Error(), "not found")
}

func TestShipmentRepo_Update(t *testing.T) {
db, mock := newMockDB(t)
repo := NewShipmentRepository(db)
s := &domain.Shipment{ID: uuid.New(), SenderID: uuid.New(), Status: domain.StatusPickedUp}
mock.ExpectBegin()
mock.ExpectExec(regexp.QuoteMeta(`UPDATE "shipments"`)).WillReturnResult(sqlmock.NewResult(1, 1))
mock.ExpectCommit()
assert.NoError(t, repo.Update(context.Background(), s))
}

func TestShipmentRepo_AddEvent(t *testing.T) {
db, mock := newMockDB(t)
repo := NewShipmentRepository(db)
ev := &domain.ShipmentEvent{ID: uuid.New(), ShipmentID: uuid.New(), Status: domain.StatusPickedUp, RecordedBy: uuid.New()}
mock.ExpectBegin()
mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shipment_events"`)).
WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(time.Now()))
mock.ExpectCommit()
assert.NoError(t, repo.AddEvent(context.Background(), ev))
}

func TestShipmentRepo_ListBySender(t *testing.T) {
db, mock := newMockDB(t)
repo := NewShipmentRepository(db)
senderID := uuid.New()
mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows(nil))
shipments, total, err := repo.ListBySender(context.Background(), senderID, 10, 0)
assert.NoError(t, err)
assert.Equal(t, int64(0), total)
assert.Empty(t, shipments)
}

func TestShipmentRepo_ListBySender_CountError(t *testing.T) {
db, mock := newMockDB(t)
repo := NewShipmentRepository(db)
senderID := uuid.New()
mock.ExpectQuery(`SELECT`).WillReturnError(fmt.Errorf("count error"))
_, _, err := repo.ListBySender(context.Background(), senderID, 10, 0)
require.Error(t, err)
}

func TestShipmentRepo_FindByID_Found(t *testing.T) {
db, mock := newMockDB(t)
repo := NewShipmentRepository(db)
id := uuid.New()
senderID := uuid.New()
mock.ExpectQuery(`SELECT`).
WithArgs(id, 1).
WillReturnRows(sqlmock.NewRows([]string{"id", "sender_id", "tracking_number", "status", "recipient_name"}).
AddRow(id, senderID, "NX-x", "PENDING", "Test"))
mock.ExpectQuery(`SELECT`).
WithArgs(id).
WillReturnRows(sqlmock.NewRows(nil))
s, err := repo.FindByID(context.Background(), id)
require.NoError(t, err)
assert.Equal(t, id, s.ID)
}

func TestShipmentRepo_FindByTrackingNumber_Found(t *testing.T) {
db, mock := newMockDB(t)
repo := NewShipmentRepository(db)
id := uuid.New()
tn := "NX-found123"
mock.ExpectQuery(`SELECT`).
WithArgs(tn, 1).
WillReturnRows(sqlmock.NewRows([]string{"id", "tracking_number", "status"}).
AddRow(id, tn, "PENDING"))
mock.ExpectQuery(`SELECT`).
WithArgs(id).
WillReturnRows(sqlmock.NewRows(nil))
s, err := repo.FindByTrackingNumber(context.Background(), tn)
require.NoError(t, err)
assert.Equal(t, id, s.ID)
}

func TestUserRepo_FindByEmail_Found(t *testing.T) {
db, mock := newMockDB(t)
repo := NewUserRepository(db)
id := uuid.New()
mock.ExpectQuery(`SELECT`).
WithArgs("admin@nexus.com", 1).
WillReturnRows(sqlmock.NewRows([]string{"id", "email", "role", "active"}).
AddRow(id, "admin@nexus.com", "admin", true))
u, err := repo.FindByEmail(context.Background(), "admin@nexus.com")
require.NoError(t, err)
assert.Equal(t, id, u.ID)
}

func TestUserRepo_FindByID_Found(t *testing.T) {
db, mock := newMockDB(t)
repo := NewUserRepository(db)
id := uuid.New()
mock.ExpectQuery(`SELECT`).
WithArgs(id, 1).
WillReturnRows(sqlmock.NewRows([]string{"id", "email", "role", "active"}).
AddRow(id, "user@nexus.com", "operator", true))
u, err := repo.FindByID(context.Background(), id)
require.NoError(t, err)
assert.Equal(t, id, u.ID)
}

func TestUserRepo_New(t *testing.T) {
db, _ := newMockDB(t)
assert.NotNil(t, NewUserRepository(db))
}

func TestUserRepo_Create(t *testing.T) {
db, mock := newMockDB(t)
repo := NewUserRepository(db)
u := &domain.User{ID: uuid.New(), Email: "t@nexus.com", PasswordHash: "h", Role: domain.RoleOperator, Active: true}
mock.ExpectBegin()
mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(time.Now(), time.Now()))
mock.ExpectCommit()
assert.NoError(t, repo.Create(context.Background(), u))
}

func TestUserRepo_FindByEmail_NotFound(t *testing.T) {
db, mock := newMockDB(t)
repo := NewUserRepository(db)
mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WithArgs("ghost@nexus.com", 1).WillReturnRows(sqlmock.NewRows(nil))
_, err := repo.FindByEmail(context.Background(), "ghost@nexus.com")
require.Error(t, err)
assert.Contains(t, err.Error(), "not found")
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
db, mock := newMockDB(t)
repo := NewUserRepository(db)
id := uuid.New()
mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WithArgs(id, 1).WillReturnRows(sqlmock.NewRows(nil))
_, err := repo.FindByID(context.Background(), id)
require.Error(t, err)
assert.Contains(t, err.Error(), "not found")
}

func TestUserRepo_UpdateLastLogin(t *testing.T) {
db, mock := newMockDB(t)
repo := NewUserRepository(db)
id := uuid.New()
mock.ExpectBegin()
mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users"`)).
WithArgs(sqlmock.AnyArg(), id).
WillReturnResult(sqlmock.NewResult(1, 1))
mock.ExpectCommit()
assert.NoError(t, repo.UpdateLastLogin(context.Background(), id))
}
