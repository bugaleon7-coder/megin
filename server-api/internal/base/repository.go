package base

import (
	"database/sql"
	"errors"
	"fmt"
	commonDto "megin/internal/module/common/dto"
	"time"

	"megin/internal/config"
	"megin/pkg/context/api"

	"gorm.io/gorm"
)

type IRepo interface {
	GetById(id int) (Model, error)
}

type TX struct {
	*gorm.DB
}

func TxBegin(opts ...*sql.TxOptions) *TX {
	db := config.GetMysqlDB().Begin(opts...)
	return &TX{
		DB: db,
	}
}

// 自动处理事务的提交和回滚,参考测试用例
// ArticleCreateTx(tx *repo.TX) (err error)
func (tx *TX) AutoCommitHandler(err *error) {
	if err != nil {
		rollError := tx.Rollback()
		if rollError != nil {
			*err = fmt.Errorf("business error: %w, rollback failed: %v", *err, rollError)
		}
		return
	}
	commitErr := tx.Commit()
	if commitErr != nil {
		*err = fmt.Errorf("commit failed: %w", commitErr)
	}
}

// Commit 提交事务（重写以处理 nil）
func (tx *TX) Commit() error {
	if tx == nil || tx.DB == nil {
		return errors.New("tx is nil")
	}
	return tx.DB.Commit().Error
}

// Rollback 回滚事务（重写以处理 nil）
func (tx *TX) Rollback() error {
	if tx == nil || tx.DB == nil {
		return errors.New("tx is nil")
	}
	return tx.DB.Rollback().Error
}

// ToGormDB 转换为 *gorm.DB（如果需要）
func (tx *TX) ToGormDB() *gorm.DB {
	if tx == nil {
		return nil
	}
	return tx.DB
}

type Repository[T Model] struct {
	ctx *api.Context
	db  *gorm.DB //默认连接
}

func (this *Repository[T]) DB() *gorm.DB {
	return this.db
}

func (this *Repository[T]) EnableTx(tx *TX) *Repository[T] {
	this.db = tx.DB
	return this
}

func (this *Repository[T]) Commit() error {
	if this.db == nil {
		return errors.New("事务未声明,请先调用Begin()")
	}
	this.db.Commit()
	return nil
}

func (this *Repository[T]) Rollback() error {
	if this.db == nil {
		return errors.New("事务未声明,请先调用Begin()")
	}
	this.db.Rollback()
	return nil
}

func (this *Repository[T]) GetById(id any) (T, error) {
	var model T
	err := this.db.Where("id", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model, nil
	}
	return model, err
}

func (this *Repository[T]) DeleteById(id any) error {
	var model T
	err := this.db.Where("id", id).Delete(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (this *Repository[T]) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	err := this.db.Transaction(fc, opts...)
	return err
}

func (this *Repository[T]) Create(m Model) error {
	err := this.db.Create(m).Error
	return err
}

func (this *Repository[T]) Save(m Model) error {
	err := this.db.Save(m).Error
	return err
}

func (this *Repository[T]) Initialize(ctx *api.Context) {
	if ctx.Tx != nil {
		this.db = ctx.Tx
	} else {
		this.db = config.GetMysqlDB()
	}
	this.ctx = ctx
}

func (this *Repository[T]) GetLastInsertID() (int, error) {
	var id int
	err := this.db.Raw("SELECT LAST_INSERT_ID()").Scan(&id).Error
	return id, err
}

func (this *Repository[T]) GetRows(where, order string, limit int) ([]T, error) {
	var rows []T
	query := this.db.Where(where).Order(order)
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&rows).Error
	return rows, err
}

// BuildQueryConditions 构建通用查询条件
func BuildQueryConditions(query *gorm.DB, filter commonDto.QueryFilter, keywordFields []string, timeField string) *gorm.DB {
	// 关键词搜索
	if filter.Keyword != "" && len(keywordFields) > 0 {
		query = query.Where(func(db *gorm.DB) *gorm.DB {
			for i, field := range keywordFields {
				if i == 0 {
					db = db.Where(field+" LIKE ?", "%"+filter.Keyword+"%")
				} else {
					db = db.Or(field+" LIKE ?", "%"+filter.Keyword+"%")
				}
			}
			return db
		})
	}

	// 时间范围查询
	if filter.StartTime != "" && timeField != "" {
		if startTime, err := time.Parse("2006-01-02", filter.StartTime); err == nil {
			query = query.Where(timeField+" >= ?", startTime.Unix())
		}
	}
	if filter.EndTime != "" && timeField != "" {
		if endTime, err := time.Parse("2006-01-02", filter.EndTime); err == nil {
			// 设置为当天的最后一秒
			endTime = endTime.Add(24*time.Hour - time.Second)
			query = query.Where(timeField+" <= ?", endTime.Unix())
		}
	}

	// 排序
	if filter.SortField != "" {
		order := filter.SortField
		if filter.SortOrder == "" {
			filter.SortOrder = "desc"
		}
		query = query.Order(order + " " + filter.SortOrder)
	}

	return query
}

func PageQuery[T Model](query *gorm.DB, page commonDto.PageQuery) (*commonDto.PageResult[T], error) {
	result := new(commonDto.PageResult[T])
	if page.PageNo == 0 {
		page.PageNo = 1
	}

	if page.PageSize == 0 {
		page.PageSize = 20
	}

	result.PageNo = page.PageNo
	result.PageSize = page.PageSize
	offset := (page.PageNo - 1) * page.PageSize
	var rows []T
	err := query.Count(&result.TotalSize).Error
	if err != nil {
		return nil, err
	}

	err = query.Limit(page.PageSize).Offset(offset).Find(&rows).Error
	if (int(result.TotalSize) % page.PageSize) > 0 {
		result.TotalPage = result.TotalSize/int64(page.PageSize) + 1
	} else {
		result.TotalPage = result.TotalSize / int64(page.PageSize)
	}
	result.List = rows
	return result, err
}
