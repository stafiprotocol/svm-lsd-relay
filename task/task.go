package task

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"svm-lsd-relay/pkg/config"
	"svm-lsd-relay/pkg/lsd_program"
	"svm-lsd-relay/pkg/utils"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type Task struct {
	stop           chan struct{}
	cfg            config.ConfigStart
	lsdProgramID   solana.PublicKey
	stakingProgram solana.PublicKey
	stakeManager   solana.PublicKey

	stakingTokenMint solana.PublicKey
	lsdTokenMint     solana.PublicKey

	tokenProgramId solana.PublicKey

	feePayerAccount solana.PrivateKey

	client   *rpc.Client
	handlers []Handler
}

type Handler struct {
	method func(*lsd_program.StakeManager) error
	name   string
}

func NewTask(cfg config.ConfigStart, feePayer solana.PrivateKey) *Task {
	s := &Task{
		stop:            make(chan struct{}),
		cfg:             cfg,
		feePayerAccount: feePayer,
	}
	return s
}

func (task *Task) Start() error {
	rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
		task.cfg.RpcEndpoint,
		rate.Every(time.Second), // time frame
		5,                       // limit of requests per time frame
	))
	task.client = rpcClient

	task.lsdProgramID = solana.MustPublicKeyFromBase58(task.cfg.LsdProgramID)
	task.stakeManager = solana.MustPublicKeyFromBase58(task.cfg.StakeManagerAddress)
	stakeManager, err := utils.GetSvmLsdStakeManager(task.client, task.stakeManager)
	if err != nil {
		return err
	}

	task.lsdTokenMint = stakeManager.LsdTokenMint
	task.stakingTokenMint = stakeManager.StakingTokenMint
	task.stakingProgram = stakeManager.StakingProgram

	stakingTokenMintAccount, err := rpcClient.GetAccountInfo(context.Background(), task.stakingTokenMint)
	if err != nil {
		return err
	}
	task.tokenProgramId = stakingTokenMintAccount.Value.Owner

	lsd_program.SetProgramID(task.lsdProgramID)

	task.appendHandlers(task.EraNew, task.EraWithdraw, task.EraBond,task.EraUnbond, task.EraActive)

	SafeGoWithRestart(task.handler)
	return nil
}

func (task *Task) Stop() {
	close(task.stop)
}

func (s *Task) appendHandlers(handlers ...func(*lsd_program.StakeManager) error) {
	for _, handler := range handlers {

		funcNameRaw := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()

		splits := strings.Split(funcNameRaw, "/")
		funcName := splits[len(splits)-1]
		funcName = strings.Split(funcName, ".")[2]
		funcName = strings.Split(funcName, "-")[0]

		s.handlers = append(s.handlers, Handler{
			method: handler,
			name:   funcName,
		})
	}
}

func (s *Task) handler() {
	logrus.Info("start handlers")
	retry := 0

	for {
		if retry > 200 {
			utils.ShutdownRequestChannel <- struct{}{}
			return
		}
		select {
		case <-s.stop:
			logrus.Info("task has stopped")
			return
		default:
			err := s.handleEra()
			if err != nil {
				logrus.Warnf("era handle failed: %s, will retry.", err)
				time.Sleep(time.Second * 6)
				retry++
				continue
			}

			retry = 0
		}

		time.Sleep(10 * time.Second)
	}
}

func (t *Task) handleEra() error {
	for _, handler := range t.handlers {
		funcName := handler.name
		logrus.Debugf("handler %s start...", funcName)
		stakeManager, err := utils.GetSvmLsdStakeManager(t.client, t.stakeManager)
		if err != nil {
			return err
		}
		err = handler.method(stakeManager)
		if err != nil {
			return fmt.Errorf("handler %s failed: %s, will retry", funcName, err)
		}
		logrus.Debugf("handler %s end", funcName)
	}

	return nil
}
