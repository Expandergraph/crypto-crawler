package collector

import (
	"encoding/json"
	"fmt"

	"github.com/Expandergraph/crypto-crawler/model"
	"github.com/Expandergraph/crypto-crawler/rpc"
)

func GetFullBlockByNumber() {
	requester := rpc.NewETHRPCRequester(nodeUrl)

	// 获取区块号
	number, _ := requester.GetLatestBlockNumber()

	// 获取区块信息
	fullBlock, err := requester.GetBlockInfoByNumber(number)

	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("获取区块信息失败，信息是：", err.Error())
		return
	}
	// 查询成功，将 区块 结果的结构体以 json 格式序列化，再以 string 格式输出
	json1, _ := json.Marshal(fullBlock)
	fmt.Println("根据区块号获取区块信息", string(json1))

	// 根据区块 hash 获取区块信息
	fullBlock, err = requester.GetBlockInfoByHash(fullBlock.ParentHash)
	json2, _ := json.Marshal(fullBlock)
	fmt.Println("根据区块hash获取区块信息", string(json2))

	logs, err := requester.EthGetLogs(model.FilterParams{
		FromBlock: fullBlock.Number,
		ToBlock:   fullBlock.Number,
	})
	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("获取log信息失败，信息是：", err.Error())
		return
	}

	fmt.Println(logs)
}

/*
            token_transfer = EthTokenTransfer()
            token_transfer.token_address = to_normalized_address(receipt_log.address)
            token_transfer.from_address = word_to_address(topics_with_data[1])
            token_transfer.to_address = word_to_address(topics_with_data[2])
            token_transfer.value = hex_to_dec(topics_with_data[3])
            token_transfer.transaction_hash = receipt_log.transaction_hash
            token_transfer.log_index = receipt_log.log_index
            token_transfer.block_number = receipt_log.block_number
            return token_transfer

        return None


def split_to_words(data):
    if data and len(data) > 2:
        data_without_0x = data[2:]
        words = list(chunk_string(data_without_0x, 64))
        words_with_0x = list(map(lambda word: '0x' + word, words))
        return words_with_0x
    return []


def word_to_address(param):
    if param is None:
        return None
    elif len(param) >= 40:
        return to_normalized_address('0x' + param[-40:])
    else:
        return to_normalized_address(param)

*/
