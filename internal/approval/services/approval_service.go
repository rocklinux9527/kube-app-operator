package services

import (
	"fmt"
	commontype "github.com/k8s/kube-app-operator/internal/api/types"
	"github.com/k8s/kube-app-operator/internal/approval/models"
	"github.com/k8s/kube-app-operator/internal/approval/repositories"
	"github.com/k8s/kube-app-operator/internal/custom/extendLogic"
	"time"
)

type RequestService struct {
    repo *repositories.RequestRepo
	userRepo *repositories.UserRepo
}

func NewRequestService(repo *repositories.RequestRepo, userRepo *repositories.UserRepo) *RequestService {
    return &RequestService{repo: repo,userRepo: userRepo}
}

// -------------------- 输入结构 --------------------

type CreateRequestInput struct {
    CreatedBy    string `json:"created_by"`
    BusinessLine string `json:"business_line"`
    ServiceName  string `json:"service_name"`
    Image        string `json:"image"`
    Replicas     int    `json:"replicas"`
    TemplateName string `json:"template_name"`
    Purpose      string `json:"purpose"`
    Operation    string `json:"operation,omitempty"`
}



type DeleteRequestInput struct {
	BusinessLine string `json:"business_line" binding:"required"`
	ServiceName  string `json:"service_name" binding:"required"`
	CreatedBy    string `json:"created_by" binding:"required"`
	Purpose      string `json:"purpose,omitempty"` // 可以传说明，选填
}

type ApprovalInput struct {
    ApproverRole string `json:"approver_role"`
    ApproverName string `json:"approver_name"`
    Decision     string `json:"decision"` // APPROVE / REJECT
    Comment      string `json:"comment"`
}

// -------------------- 用户角色校验 --------------------

// 检查用户是否属于某个角色

func (s *RequestService) checkUserRole(userName, role string) error {
	ok, err := s.userRepo.UserHasRole(userName, role)
	if err != nil {
		return fmt.Errorf("校验用户角色失败: %w", err)
	}
	if !ok {
		return fmt.Errorf("用户 [%s] 不具备角色 [%s]，无权执行审批", userName, role)
	}
	return nil
}

func (s *RequestService) CreateRequest(input CreateRequestInput) (*models.Request, error) {
     op := input.Operation
     if op == "" {
        op = "CREATE"
     }
    req := &models.Request{
        CreatedBy:    input.CreatedBy,
        BusinessLine: input.BusinessLine,
        ServiceName:  input.ServiceName,
        Image:        input.Image,
        Replicas:     input.Replicas,
        TemplateName: input.TemplateName,
        Purpose:      input.Purpose,
        Status:       "PENDING",
        Operation:    op,
    }

    if err := s.repo.Create(req); err != nil {
        return nil, err
    }
    return req, nil
}


// -------------------- 删除请求 --------------------

func (s *RequestService) DeleteRequest(input DeleteRequestInput) error {
	req := &models.Request{
		BusinessLine: input.BusinessLine,
		ServiceName:  input.ServiceName,
	}
	return s.repo.Delete(req)
}


func (s *RequestService) ApproveRequest(requestID string, input ApprovalInput) (*models.Request, error) {
    req, err := s.repo.FindByRequestID(requestID)
    if err != nil {
      //  return nil, err
        return nil,fmt.Errorf("审批失败：未找到请求 %s，错误: %w", requestID, err)
    }

	// 校验用户角色

	if err := s.checkUserRole(input.ApproverName, input.ApproverRole); err != nil {
		return nil, err
	}

	newStatus, valid := getNextStatus(req.Status, input.ApproverRole, input.Decision)
    if !valid {
        //return nil, errors.New("invalid transition")
        return nil, fmt.Errorf("审批失败：当前审批阶段是 [%s]，不允许由角色 [%s] 执行决策 [%s]", req.Status, input.ApproverRole, input.Decision)
    }

    // 使用 Repo.WithTx，而不是 Transaction
    err = s.repo.WithTx(func(txRepo *repositories.RequestRepo) error {
        req.Status = newStatus
        req.LastUpdated = time.Now()
        if err := txRepo.Update(req); err != nil {
            return err
        }

        approval := &models.Approval{
            RequestID:    req.RequestID,
            ApproverRole: input.ApproverRole,
            ApproverName: input.ApproverName,
            Decision:     input.Decision,
            Comment:      input.Comment,
        }
        if err := txRepo.CreateApproval(approval); err != nil {
            return err
        }

        history := &models.RequestHistory{
            RequestID: req.RequestID,
            Status:    newStatus,
            ChangedBy: input.ApproverName,
            Note:      input.Comment,
        }
        if err := txRepo.CreateHistory(history); err != nil {
            return err
        }

        return nil
    })
    if err != nil {
        return nil, err
    }

    // Repo 内部 Update 会异步更新 Redis（无双写）

    if err := s.repo.Update(req); err != nil {
        return nil, err
    }

    // 模拟部署逻辑
    if req.Status == "K8S_APPROVED" {
       wrapAndDeploy(s,req)
    }

    return req, nil
}



// wrapAndDeploy 根据请求的 Operation 调用不同的 K8s 操作

func wrapAndDeploy(s *RequestService, req *models.Request) {
    go func(r *models.Request) {
        // 模拟一点延迟（可选）
        time.Sleep(1 * time.Second)
        switch r.Operation {
        case "CREATE":
            appReq := commontype.KubeAppRequest{
                Name:         r.ServiceName,
                Namespace:    r.BusinessLine,
                Image:        r.Image,
                Replicas:     int32(r.Replicas),
                TemplateType: r.TemplateName,
				TemplateName: r.TemplateName,
            }
            if err := extendLogic.InternalCreateKubeApp(appReq); err != nil {
                fmt.Println("❌ Failed to create KubeApp:", err.Error())
                return
            }
            fmt.Println("🚀 KubeApp created successfully:", r.ServiceName)
        case "UPDATE":
            appReq := commontype.KubeAppRequest{
                Name:         r.ServiceName,
                Namespace:    r.BusinessLine,
                Image:        r.Image,
                Replicas:     int32(r.Replicas),
                TemplateType: r.TemplateName,
				TemplateName: r.TemplateName,
            }
			fmt.Println("更新operator参数", r.ServiceName,r.BusinessLine,r.Image,r.Replicas)
            if err := extendLogic.InternalUpdateKubeApp(appReq); err != nil {
                fmt.Println("❌ Failed to update KubeApp:", err.Error())
                return
            }
            fmt.Println("🔄 KubeApp updated successfully:", r.ServiceName)

        case "DELETE":
            delReq := commontype.KubeDeleteAppRequest{
                Name:      r.ServiceName,
                Namespace: r.BusinessLine,
                DeleteKubeApp: true,
            }
            if err := extendLogic.InternalDeleteKubeApp(delReq); err != nil {
                fmt.Println("❌ Failed to delete KubeApp:", err.Error())
                return
            }

            fmt.Println("🗑️ KubeApp deleted successfully:", r.ServiceName)

			if err := s.repo.Delete(r); err != nil {
				fmt.Println("❌ Failed to delete DB record:", err.Error())
				return
			}
		default:
            fmt.Println("⚠️ Unsupported operation:", r.Operation)
        }
    }(req)
}



// -------------------- RejectRequest --------------------

func (s *RequestService) RejectRequest(requestID string, input ApprovalInput) (*models.Request, error) {
	// 查询请求
	req, err := s.repo.FindByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("拒绝失败：未找到请求 %s，错误: %w", requestID, err)
	}

	// 校验用户角色
	if err := s.checkUserRole(input.ApproverName, input.ApproverRole); err != nil {
		return nil, err
	}

	// 校验是否允许 REJECT
	newStatus, ok := getNextStatus(req.Status, input.ApproverRole, "REJECT")
	if !ok {
		return nil, fmt.Errorf(
			"拒绝失败：当前审批阶段是 [%s]，不允许由角色 [%s] 执行 REJECT 操作",
			req.Status, input.ApproverRole,
		)
	}
	// 事务内写入（Update + Approval + History）
	err = s.repo.WithTx(func(txRepo *repositories.RequestRepo) error {
		req.Status = newStatus
		req.LastUpdated = time.Now()
		if err := txRepo.Update(req); err != nil {
			return err
		}

		approval := &models.Approval{
			RequestID:    req.RequestID,
			ApproverRole: input.ApproverRole,
			ApproverName: input.ApproverName,
			Decision:     "REJECT",
			Comment:      input.Comment,
			CreatedAt:    time.Now(),
		}
		if err := txRepo.CreateApproval(approval); err != nil {
			return err
		}

		history := &models.RequestHistory{
			RequestID: req.RequestID,
			Status:    newStatus,
			ChangedBy: input.ApproverName,
			Note:      input.Comment,
			CreatedAt: time.Now(),
		}
		if err := txRepo.CreateHistory(history); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("拒绝事务执行失败: %w", err)
	}

	return req, nil
}

// -------------------- 分页查询 --------------------

func (s *RequestService) ListRequests(page, pageSize int) ([]models.Request, int64, error) {
    return s.repo.ListRequests(page, pageSize)
}

// -------------------- 批量查询 --------------------

func (s *RequestService) BatchFindByIDs(ids []string) ([]models.Request, error) {
    return s.repo.BatchFindByIDs(ids)
}

func (s *RequestService) FindByID(id string) (*models.Request, error) {
	return s.repo.FindByID(id)
}



// -------------------- 状态流转规则 --------------------
func getNextStatus(current, role, decision string) (string, bool) {
    switch current {
    case "PENDING":
        if role == "OPS" && decision == "APPROVE" {
            return "OPS_APPROVED", true
        }
        if role == "OPS" && decision == "REJECT" {
            return "OPS_REJECTED", true
        }
    case "OPS_APPROVED":
        if role == "SRE" && decision == "APPROVE" {
            return "SRE_APPROVED", true
        }
        if role == "SRE" && decision == "REJECT" {
            return "SRE_REJECTED", true
        }
    case "SRE_APPROVED":
        if role == "K8S" && decision == "APPROVE" {
            return "K8S_APPROVED", true
        }
        if role == "K8S" && decision == "REJECT" {
            return "K8S_REJECTED", true
        }
    }
    return "", false
}

