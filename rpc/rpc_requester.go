package rpc

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/Expandergraph/crypto-crawler/model"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

type ETHRPCRequester struct {
	nonceManager *NonceManager // nonce 管理者实例
	client       *ETHRPCClient // 小写字母开头，私有的 rpc 客户端实例
}

func NewETHRPCRequester(nodeUrl string) *ETHRPCRequester {
	requester := &ETHRPCRequester{}
	// 实例化 nonce 管理器
	requester.nonceManager = NewNonceManager()
	// 实例化 rpc 客户端
	requester.client = NewETHRPCClient(nodeUrl)
	return requester
}

// 根据交易的 hash 值获取对应交易的信息
func (r *ETHRPCRequester) GetTransactionByHash(txHash string) (model.Transaction, error) {
	methodName := "eth_getTransactionByHash"
	result := model.Transaction{}
	// 下面 call 的 result 传入的是  model.Transaction 结构体的引用，
	// 这样内部所设置的值在函数执行完之后才能依然有效果
	err := r.client.GetRpc().Call(&result, methodName, txHash)
	return result, err
}

// 根据交易 hash 字符串数组批量获取对应的交易信息
func (r *ETHRPCRequester) GetTransactions(txHashs []string) ([]*model.Transaction, error) {
	name := "eth_getTransactionByHash"

	// 结果数组存储的是每个请求的结果指针，也就是引用
	rets := []*model.Transaction{}

	// 获取 hash 数组的长度，方便在循环中逐个实例化 BatchElem
	size := len(txHashs)

	reqs := []rpc.BatchElem{}
	for i := 0; i < size; i++ {
		ret := model.Transaction{}
		// 实例化每个 BatchElem
		req := rpc.BatchElem{
			Method: name,
			Args:   []interface{}{txHashs[i]},
			// &ret 传入单个请求的结果引用，这样是保证它在函数内部被修改值后，回到函数外来，值仍有效
			Result: &ret,
		}
		reqs = append(reqs, req)  // 将每个 BatchElem 添加到 BatchElem 数组
		rets = append(rets, &ret) // 每个请求的结果引用添加到结果数组中
	}
	err := r.client.GetRpc().BatchCall(reqs) // 传入 BatchElem 数组，发起批量请求
	return rets, err
}

// 单条查询：根据以太坊地址，查询以太坊 eth 的余额
func (r *ETHRPCRequester) GetETHBalance(address string) (string, error) {
	name := "eth_getBalance"
	result := ""
	// 对应文档，第一个参数就是要被查询的以太坊地址
	// 第二个参数就是 latest
	err := r.client.GetRpc().Call(&result, name, address, "latest")
	if err != nil {
		return "", err
	}
	if result == "" {
		return "", errors.New("eth balance is null")
	}
	// 因为查询所返回的结果是一个16进制的字符串，
	// 为了方便阅读，我们在下面使用 go 的大数处理将其转为 10 进制，
	// 并防止数位溢出
	ten, _ := new(big.Int).SetString(result[2:], 16)
	return ten.String(), nil
}

// 批量查询：根据以太坊地址数组，查询以太坊 eth 的余额
func (r *ETHRPCRequester) GetETHBalances(addresss []string) ([]string, error) {
	name := "eth_getBalance"
	// 结果数组存储的是每个请求的结果指针，也就是引用
	rets := []*string{}
	// 获取 addresss 数组的长度，方便在循环中逐个实例化 BatchElem
	size := len(addresss)
	reqs := []rpc.BatchElem{}
	for i := 0; i < size; i++ {
		ret := ""
		// 实例化每个 BatchElem
		req := rpc.BatchElem{
			Method: name,
			Args:   []interface{}{addresss[i], "latest"},
			// &ret 传入单个请求的结果引用，这样是保证它在函数内部被修改值后，回到函数外来，值仍有效
			Result: &ret,
		}
		reqs = append(reqs, req)  // 将每个 BatchElem 添加到 BatchElem 数组
		rets = append(rets, &ret) // 每个请求的结果引用添加到结果数组中
	}
	err := r.client.GetRpc().BatchCall(reqs) // 传入 BatchElem 数组，发起批量请求
	if err != nil {
		return nil, err
	}
	// 查询每个请求有没有错误
	for _, req := range reqs {
		if req.Error != nil {
			return nil, req.Error // 返回错误
		}
	}
	finalRet := []string{}
	for _, item := range rets {
		ten, _ := new(big.Int).SetString((*item)[2:], 16)
		finalRet = append(finalRet, ten.String())
	}
	return finalRet, err
}

// ERC20BalanceRpcReq 是查询 ERC20 代币的参数集合结构体
type ERC20BalanceRpcReq struct {
	ContractAddress string // 合约的以太坊地址
	UserAddress     string // 用户的以太坊地址
	ContractDecimal int    // 合约所对应代币的数位
}

// 批量查询：根据以太坊地址数组，查询 ERC20 代币的余额
func (r *ETHRPCRequester) GetERC20Balances(paramArr []ERC20BalanceRpcReq) ([]string, error) {
	name := "eth_call"
	methodId := "0x70a08231" // 这个就是 balanceOf 的 methodId
	// 结果数组存储的是每个请求的结果指针，也就是引用
	rets := []*string{}
	// 获取参数数组的长度，方便在循环中逐个实例化 BatchElem
	size := len(paramArr)
	reqs := []rpc.BatchElem{}

	for i := 0; i < size; i++ {
		ret := ""
		arg := &model.CallArg{}
		userAddress := paramArr[i].UserAddress
		// 下面是针对访问 balanceOf 时的必须参数，查询余额是不需要油费的，但是发现一些版本的节点又需要指定这个参数，所以下面还是指定一个
		arg.Gas = hexutil.EncodeUint64(300000)
		arg.To = common.HexToAddress(paramArr[i].ContractAddress)
		//  data 参数的组合格式见 “交易参数的说明” 小节中的详解
		arg.Data = methodId + "000000000000000000000000" + userAddress[2:]
		// 实例化每个 BatchElem
		req := rpc.BatchElem{
			Method: name,
			Args:   []interface{}{arg, "latest"},
			// &ret 传入单个请求的结果引用，这样是保证它在函数内部被修改值后，回到函数外来，值仍有效
			Result: &ret,
		}
		reqs = append(reqs, req)  // 将每个 BatchElem 添加到 BatchElem 数组
		rets = append(rets, &ret) // 每个请求的结果引用添加到结果数组中
	}
	err := r.client.GetRpc().BatchCall(reqs) // 传入 BatchElem 数组，发起批量请求
	if err != nil {
		return nil, err
	}
	// 查询每个请求有没有错误
	for _, req := range reqs {
		if req.Error != nil {
			return nil, req.Error // 返回错误
		}
	}
	finalRet := []string{}
	for _, item := range rets {
		if *item == "" {
			continue
		}
		ten, _ := new(big.Int).SetString((*item)[2:], 16)
		finalRet = append(finalRet, ten.String())
	}
	return finalRet, err
}

// 创建以太坊钱包
func (r *ETHRPCRequester) CreateETHWallet(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cant empty")
	}
	if len(password) < 6 {
		return "", errors.New("password's len must more than 6 words")
	}
	keydir := "./keystores" // 它是用来存储所创建了的钱包的 keystore 文件的目录
	// StandardScryptN 是Scrypt加密算法的标准N参数
	// StandardScryptP 是Scrypt加密算法的标准P参数
	ks := keystore.NewKeyStore(keydir, keystore.StandardScryptN, keystore.StandardScryptP)
	wallet, err := ks.NewAccount(password) // 传入密码，创建钱包
	if err != nil {
		return "0x", err
	}
	return wallet.Address.String(), nil
}

// 获取地址的 nonce 值
func (r *ETHRPCRequester) GetNonce(address string) (uint64, error) {
	methodName := "eth_getTransactionCount"
	nonce := ""
	// 因为我们要查询最新的，根据基于 etTransactionCount 情况下的区块号关系，选取 pending
	err := r.client.client.Call(&nonce, methodName, address, "pending")
	if err != nil {
		return 0, fmt.Errorf("发送交易失败! %s", err.Error())
	}
	n, _ := new(big.Int).SetString(nonce[2:], 16) // 采用大数类型将 16 进制的结果转为 10 进制
	return n.Uint64(), nil                        // 返回交易 hash
}

// 获取以太坊最新生成区块的区块号
func (r *ETHRPCRequester) GetLatestBlockNumber() (*big.Int, error) {
	methodName := "eth_blockNumber"
	number := ""
	// eth_blockNumber 不需要参数
	err := r.client.client.Call(&number, methodName)
	if err != nil {
		return nil, fmt.Errorf("获取最新区块号失败! %s", err.Error())
	}
	ten, _ := new(big.Int).SetString(number[2:], 16)
	return ten, nil
}

// 根据区块号，获取区块信息
func (r *ETHRPCRequester) GetBlockInfoByNumber(blockNumber *big.Int) (*model.FullBlock, error) {
	number := fmt.Sprintf("%#x", blockNumber)
	methodName := "eth_getBlockByNumber"
	fullBlock := model.FullBlock{}
	// eth_getBlockByNumber 的第二个参数：
	// 如果是 true，则返回完整的区块信息，false 则 transaction 部分只返回交易hash数组
	err := r.client.client.Call(&fullBlock, methodName, number, true)
	if err != nil {
		return nil, fmt.Errorf("get block info failed! %s", err.Error())
	}
	if fullBlock.Number == "" {
		return nil, fmt.Errorf("block info is empty %s", blockNumber.String())
	}
	return &fullBlock, nil
}

// 根据区块 hash，获取区块信息
func (r *ETHRPCRequester) GetBlockInfoByHash(blockHash string) (*model.FullBlock, error) {
	methodName := "eth_getBlockByHash"
	fullBlock := model.FullBlock{}
	// eth_getBlockByHash 的第二个参数：
	// 如果是 true，则返回完整的区块信息，false 则 transaction 部分只返回交易hash数组
	err := r.client.client.Call(&fullBlock, methodName, blockHash, true)
	if err != nil {
		return nil, fmt.Errorf("get block info failed! %s", err.Error())
	}
	if fullBlock.Number == "" {
		return nil, fmt.Errorf("block info is empty %s", blockHash)
	}
	return &fullBlock, nil
}

// EthGetLogs returns an array of all logs matching a given filter object.
func (r *ETHRPCRequester) EthGetLogs(params model.FilterParams) ([]model.Log, error) {
	methodName := "eth_getLogs"

	var logs []model.Log
	err := r.client.GetRpc().Call(&logs, methodName, params)
	return logs, err
}

// FilterLogs returns the logs that satisfy the supplied filter query.
//func (rc *RpcClient) FilterLogs(q ethereum.FilterQuery) ([]ethtypes.Log, error) {
//	ctx, cancel := context.WithTimeout(rc.ctx, requestTimeout)
//	defer cancel()
//	logs, err := rc.ethRPCClient.FilterLogs(ctx, q)
//	if err != nil {
//		// Add the query filter to the error so that it gets logged
//		if err.Error() == constants.ParityFilterUnknownBlock || err.Error() == constants.GethFilterUnknownBlock {
//			err = FilterUnknownBlockError{
//				Message:     err.Error(),
//				FilterQuery: q,
//			}
//		}
//		return nil, err
//	}
//	return logs, nil
//}
