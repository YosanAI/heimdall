package listener

import (
	"context"
	"errors"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	ethereum "github.com/maticnetwork/bor"
	"github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/maticnetwork/heimdall/sethu/queue"
	"github.com/tendermint/tendermint/libs/log"

	httpClient "github.com/tendermint/tendermint/rpc/client"
)

var (
	// ErrAlreadyStarted is returned when somebody tries to start an already
	// running service.
	ErrAlreadyStarted = errors.New("already started")
	// ErrAlreadyStopped is returned when somebody tries to stop an already
	// stopped service (without resetting it).
	ErrAlreadyStopped = errors.New("already stopped")
	// ErrNotStarted is returned when somebody tries to stop a not running
	// service.
	ErrNotStarted = errors.New("not started")
)

// Listener defines a block header listerner for Rootchain, Maticchain, Heimdall
type Listener interface {
	Start() error

	StartHeaderProcess(context.Context)

	StartPolling(context.Context, time.Duration)

	StartSubscription(context.Context, ethereum.Subscription)

	ProcessHeader(*types.Header)

	// PublishEvent()

	// Stop() error

	String() string

	SetLogger(log.Logger)
}

type BaseListener struct {
	Logger  log.Logger
	name    string
	started uint32 // atomic
	stopped uint32 // atomic
	quit    chan struct{}

	// The "subclass" of BaseService
	impl Listener

	// storage client
	// storageClient *leveldb.DB

	// contract caller
	contractConnector helper.ContractCaller

	// header channel
	HeaderChannel chan *types.Header

	// cancel function for poll/subscription
	cancelSubscription context.CancelFunc

	// header listener subscription
	cancelHeaderProcess context.CancelFunc

	// cli context
	cliCtx cliContext.CLIContext

	// queue connector
	queueConnector *queue.QueueConnector

	// http client to subscribe to
	httpClient *httpClient.HTTP
}

// NewBaseListener creates a new BaseListener.
func NewBaseListener(cdc *codec.Codec, queueConnector *queue.QueueConnector, logger log.Logger, name string, impl Listener) *BaseListener {

	contractCaller, err := helper.NewContractCaller()
	if err != nil {
		logger.Error("Error while getting root chain instance", "error", err)
		panic(err)
	}

	cliCtx := cliContext.NewCLIContext().WithCodec(cdc)
	cliCtx.BroadcastMode = client.BroadcastAsync
	cliCtx.TrustNode = true

	if logger == nil {
		logger = log.NewNopLogger()
	}

	// creating syncer object
	return &BaseListener{
		Logger: logger,
		name:   name,
		quit:   make(chan struct{}),
		impl:   impl,
		// storageClient: getBridgeDBInstance(viper.GetString(BridgeDBFlag)),

		cliCtx:            cliCtx,
		queueConnector:    queueConnector,
		contractConnector: contractCaller,

		HeaderChannel: make(chan *types.Header),
	}
}

// SetLogger implements Service by setting a logger.
func (bl *BaseListener) SetLogger(l log.Logger) {
	bl.Logger = l
}

// OnStart starts new block subscription
func (bl *BaseListener) Start() error {
	bl.Logger.Info("Starting listener")
	// create cancellable context
	ctx, cancelSubscription := context.WithCancel(context.Background())
	bl.cancelSubscription = cancelSubscription

	// create cancellable context
	headerCtx, cancelHeaderProcess := context.WithCancel(context.Background())
	bl.cancelHeaderProcess = cancelHeaderProcess

	// start header process
	go bl.StartHeaderProcess(headerCtx)

	// // subscribe to new head
	// subscription, err := bl.contractConnector.MainChainClient.SubscribeNewHead(ctx, bl.HeaderChannel)
	// if err != nil {
	// 	// start go routine to poll for new header using client object
	go bl.StartPolling(ctx, helper.GetConfig().SyncerPollInterval)
	// } else {
	// 	// start go routine to listen new header using subscription
	// 	go bl.StartSubscription(ctx, subscription)
	// }

	// subscribed to new head
	bl.Logger.Info("Subscribed to new head")

	return nil
}

// String implements Service by returning a string representation of the service.
func (bl *BaseListener) String() string {
	return bl.name
}

// startHeaderProcess starts header process when they get new header
func (bl *BaseListener) StartHeaderProcess(ctx context.Context) {
	bl.Logger.Info("Starting header process")
	for {
		select {
		case newHeader := <-bl.HeaderChannel:
			bl.impl.ProcessHeader(newHeader)
		case <-ctx.Done():
			return
		}
	}
}

// startPolling starts polling
func (bl *BaseListener) StartPolling(ctx context.Context, pollInterval time.Duration) {
	// How often to fire the passed in function in second
	interval := pollInterval

	bl.Logger.Info("Starting polling process")
	// Setup the ticket and the channel to signal
	// the ending of the interval
	ticker := time.NewTicker(interval)

	// start listening
	for {
		select {
		case <-ticker.C:
			header, err := bl.contractConnector.MainChainClient.HeaderByNumber(ctx, nil)
			if err == nil && header != nil {
				// send data to channel
				bl.HeaderChannel <- header
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (bl *BaseListener) StartSubscription(ctx context.Context, subscription ethereum.Subscription) {
	for {
		select {
		case err := <-subscription.Err():
			// stop service
			bl.Logger.Error("Error while subscribing new blocks", "error", err)
			// bl.Stop()

			// cancel subscription
			bl.cancelSubscription()
			return
		case <-ctx.Done():
			return
		}
	}
}
