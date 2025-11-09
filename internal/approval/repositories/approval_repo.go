
package repositories

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/k8s/kube-app-operator/internal/approval/models"
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

type RequestRepo struct {
    db  *gorm.DB
    rdb *redis.Client
}

func NewRequestRepo(db *gorm.DB, rdb *redis.Client) *RequestRepo {
    return &RequestRepo{db: db, rdb: rdb}
}

func (r *RequestRepo) redisKey(requestID string) string {
    return fmt.Sprintf("approval:request:%s", requestID)
}

// -------------------- 异步更新 Redis --------------------

func (r *RequestRepo) asyncCache(req *models.Request) {
    go func(req *models.Request) {
        ctx := context.Background()
        data, _ := json.Marshal(req)
        _ = r.rdb.Set(ctx, r.redisKey(req.RequestID), data, 10*time.Minute).Err()
    }(req)
}

// -------------------- Create --------------------

func (r *RequestRepo) Create(req *models.Request) error {
    if err := r.db.Create(req).Error; err != nil {
        return err
    }
    r.asyncCache(req)
    return nil
}

// -------------------- Update --------------------

func (r *RequestRepo) Update(req *models.Request) error {
    if err := r.db.Save(req).Error; err != nil {
        return err
    }
    r.asyncCache(req)
    return nil
}

// delete 删除表记录

func (r *RequestRepo) Delete(req *models.Request) error {
    if err := r.db.
        Where("business_line = ? AND service_name = ?", req.BusinessLine, req.ServiceName).
        Delete(&models.Request{}).Error; err != nil {
        return err
    }
    r.asyncCache(req)
    return nil
}

// -------------------- FindByRequestID --------------------

func (r *RequestRepo) FindByRequestID(requestID string) (*models.Request, error) {
    ctx := context.Background()

    // 先查 Redis
    val, err := r.rdb.Get(ctx, r.redisKey(requestID)).Result()
    if err == nil {
        var req models.Request
        if e := json.Unmarshal([]byte(val), &req); e == nil {
            return &req, nil
        }
    }

    // 再查 MySQL
    var req models.Request
    if err := r.db.Where("request_id = ?", requestID).First(&req).Error; err != nil {
        return nil, err
    }

    r.asyncCache(&req)
    return &req, nil
}

// -------------------- WithTx（事务包装，repo 版本） --------------------

func (r *RequestRepo) WithTx(fn func(txRepo *RequestRepo) error) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        txRepo := &RequestRepo{db: tx, rdb: r.rdb}
        return fn(txRepo)
    })
}

// -------------------- ListRequests 分页 --------------------

func (r *RequestRepo) ListRequests(page, pageSize int) ([]models.Request, int64, error) {
    var requests []models.Request
    var total int64

    if err := r.db.Model(&models.Request{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    offset := (page - 1) * pageSize
    if err := r.db.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&requests).Error; err != nil {
        return nil, 0, err
    }

    for _, req := range requests {
        r.asyncCache(&req)
    }

    return requests, total, nil
}

// -------------------- BatchFindByIDs --------------------

func (r *RequestRepo) BatchFindByIDs(requestIDs []string) ([]models.Request, error) {
    ctx := context.Background()
    var results []models.Request
    var missingIDs []string

    for _, id := range requestIDs {
        val, err := r.rdb.Get(ctx, r.redisKey(id)).Result()
        if err == nil {
            var req models.Request
            if e := json.Unmarshal([]byte(val), &req); e == nil {
                results = append(results, req)
                continue
            }
        }
        missingIDs = append(missingIDs, id)
    }

    if len(missingIDs) > 0 {
        var dbResults []models.Request
        if err := r.db.Where("request_id IN ?", missingIDs).Find(&dbResults).Error; err != nil {
            return nil, err
        }
        results = append(results, dbResults...)
        for _, req := range dbResults {
            r.asyncCache(&req)
        }
    }

    return results, nil
}

func (r *RequestRepo) FindByID(requestID string) (*models.Request, error) {
    ctx := context.Background()

    // 1. 先查 Redis 缓存
    val, err := r.rdb.Get(ctx, r.redisKey(requestID)).Result()
    if err == nil {
        var req models.Request
        if e := json.Unmarshal([]byte(val), &req); e == nil {
            return &req, nil
        }
    }

    // 2. 查数据库
    var req models.Request
    if err := r.db.Where("request_id = ?", requestID).First(&req).Error; err != nil {
        return nil, err
    }

    // 3. 异步写入缓存
    r.asyncCache(&req)

    return &req, nil
}


// -------------------- CreateApproval --------------------

func (r *RequestRepo) CreateApproval(appr *models.Approval) error {
    return r.db.Create(appr).Error
}

// -------------------- CreateHistory --------------------

func (r *RequestRepo) CreateHistory(hist *models.RequestHistory) error {
    return r.db.Create(hist).Error
}

// -------------------- List 别名，兼容旧 Service --------------------

func (r *RequestRepo) List(page, pageSize int) ([]models.Request, int64, error) {
    return r.ListRequests(page, pageSize)
}

