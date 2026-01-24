package cache

import (
	"context"
	"fmt"
	"log"
	"scaffolding-go/cache/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestClient_TryLock(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) redis.Cmdable

		key string

		wantErr  error
		wantLock *Lock
	}{
		{
			name: "set nx error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, context.DeadlineExceeded)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "failed to preempt lock",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: ErrFailedToPreemptLock,
		},
		{
			name: "locked",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key: "key1",
			wantLock: &Lock{
				key:        "key1",
				expiration: time.Minute,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := NewClient(tc.mock(ctrl))
			l, err := client.TryLock(context.Background(), tc.key, time.Minute)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, l.key)
			assert.Equal(t, tc.wantLock.expiration, l.expiration)
			assert.NotEmpty(t, l.value)
		})
	}
}

func TestLock_Unlock(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		key     string
		value   string
		wantErr error
	}{
		{
			name: "eval error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"value1"}).
					Return(res)
				return cmd
			},
			key:     "key1",
			value:   "value1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"value1"}).
					Return(res)
				return cmd
			},
			key:     "key1",
			value:   "value1",
			wantErr: ErrLockNotHold,
		},
		{
			name: "unlock",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"value1"}).
					Return(res)
				return cmd
			},
			key:   "key1",
			value: "value1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				client: tc.mock(ctrl),
				key:    tc.key,
				value:  tc.value,
			}
			err := lock.Unlock(context.Background())
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestLock_Refresh(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) redis.Cmdable
		key        string
		value      string
		expiration time.Duration
		wantErr    error
	}{
		{
			name: "eval error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), luaRefresh, []string{"key1"}, []any{"value1", float64(60)}).
					Return(res)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Minute,
			wantErr:    context.DeadlineExceeded,
		},
		{
			name: "lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), luaRefresh, []string{"key1"}, []any{"value1", float64(60)}).
					Return(res)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Minute,
			wantErr:    ErrLockNotHold,
		},
		{
			name: "refreshed",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), luaRefresh, []string{"key1"}, []any{"value1", float64(60)}).
					Return(res)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				client:     tc.mock(ctrl),
				key:        tc.key,
				value:      tc.value,
				expiration: tc.expiration,
			}
			err := lock.Refresh(context.Background())
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func ExampleLock_Refresh() {
	// 加锁成功，你拿到了一个lock
	var l *Lock
	stopChan := make(chan struct{})
	errChan := make(chan error)
	timeoutChan := make(chan struct{}, 1)
	timeoutRetry := 0
	go func() {
		// 间隔多久续约一次
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				// 刷新的超时时间怎么设置
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				// 出现 error 了怎么办
				err := l.Refresh(ctx)
				cancel()
				if err == context.DeadlineExceeded {
					timeoutChan <- struct{}{}
					continue
				}
				if err != nil {
					errChan <- err
					//close(stopChan)
					//close(errChan)
					return
				}
			case <-timeoutChan:
				timeoutRetry++
				if timeoutRetry > 10 {
					errChan <- context.DeadlineExceeded
					return
				}
				// 刷新的超时时间怎么设置
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				// 出现 error 了怎么办
				err := l.Refresh(ctx)
				cancel()
				if err == context.DeadlineExceeded {
					timeoutChan <- struct{}{}
					continue
				}
				if err != nil {
					errChan <- err
					//close(stopChan)
					//close(errChan)
					return
				}
				timeoutRetry = 0
			case <-stopChan:
				l.Unlock(context.Background())
				return
			}

		}
	}()

	// 这边假设就是你的业务
	// 你执行业务有很复杂的逻辑, 要记得在中间步骤检测 errChan 有没有信号
	// 循环例子——假如你的业务是循环处理
	for i := 0; i < 100; i++ {
		select {
		// 这里续约失败
		case <-errChan:
			break
		default:
			// 正常的业务逻辑
		}
	}
	// 如果没有循环，那就是每个步骤内都要检测一下
	select {
	case err := <-errChan:
		// 续约失败, 你要中断业务
		log.Fatalln(err)
		return
	default:
		// 这是你的步骤1
	}

	select {
	case err := <-errChan:
		// 续约失败, 你要中断业务
		log.Fatalln(err)
		return
	default:
		// 这是你的步骤2
	}
	select {
	case err := <-errChan:
		// 续约失败, 你要中断业务
		log.Fatalln(err)
		return
	default:
		// 这是你的步骤3
	}

	// 你的业务结束了, 就要退出续约的循环, 你怎么退出
	stopChan <- struct{}{}
}

func ExampleLock_AutoRefresh() {
	var l *Lock
	go func() {
		// 这里返回 error 你要考虑中断业务
		l.AutoRefresh(time.Second*10, time.Second)
	}()
	fmt.Println("Hello")
	// Output:
	// Hello
}
