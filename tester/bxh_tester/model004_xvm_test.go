package bxh_tester

import (
	"io/ioutil"
	"time"

	"github.com/meshplus/bitxhub-kit/hexutil"
	"github.com/meshplus/bitxhub-kit/types"
	"github.com/meshplus/bitxhub-model/pb"
	rpcx "github.com/meshplus/go-bitxhub-client"
)

//tc:部署合约，合约数据为空，交易回执状态显示失败
func (suite *Snake) Test0401_DeployContractIsNull() {
	bytes := make([]byte, 0)
	_, err := suite.client.DeployContract(bytes, nil)

	suite.Require().NotNil(err)
	suite.Require().Contains(err.Error(), "can't deploy empty contract")
}

//tc:部署合约，to地址随机，交易回执状态显示失败
func (suite *Snake) Test0402_DeployContractWithToAddress() {
	contract, err := ioutil.ReadFile("testdata/example.wasm")
	suite.Require().Nil(err)

	td := &pb.TransactionData{
		Type:    pb.TransactionData_INVOKE,
		VmType:  pb.TransactionData_XVM,
		Payload: contract,
	}

	payload, err := td.Marshal()
	suite.Require().Nil(err)
	tx := &pb.Transaction{
		From:      suite.from,
		To:        suite.to,
		Timestamp: time.Now().UnixNano(),
		Payload:   payload,
	}

	receipt, err := suite.client.SendTransactionWithReceipt(tx, nil)
	suite.Require().Nil(err)
	suite.Require().True(receipt.Status == pb.Receipt_FAILED)
	suite.Require().Contains(string(receipt.Ret), "contract byte not correct")
}

//tc:部署合约，注册部署合约，返回合约地址
func (suite *Snake) Test0403_DeployContract() {
	deployExampleContract(suite)
}

//tc:调用合约，正常调用合约，返回正确结果
func (suite *Snake) Test0404_InvokeContract() {
	address := deployExampleContract(suite)

	result, err := suite.client.InvokeXVMContract(address, "a", nil, rpcx.Int32(1), rpcx.Int32(2))
	suite.Require().Nil(err)
	suite.Require().True(result.Status == pb.Receipt_SUCCESS)
	suite.Require().True("336" == string(result.Ret))
}

//tc:调用合约，调用方法名不存在，交易回执状态显示失败
func (suite *Snake) Test0405_InvokeContractNotExistMethod() {
	address := deployExampleContract(suite)

	result, err := suite.client.InvokeXVMContract(address, "bbb", nil, rpcx.Int32(1), rpcx.Int32(2))
	suite.Require().Nil(err)
	suite.Require().True(result.Status == pb.Receipt_FAILED)
	suite.Require().Contains(string(result.Ret), "wrong rule contract")
}

//tc:调用合约，合约地址不存在，交易回执显示失败
func (suite *Snake) Test0406_InvokeRandomAddressContract() {
	// random addr len should be 42
	bs := hexutil.Encode([]byte("random contract addr"))
	fakeAddr := types.NewAddressByStr(bs)

	result, err := suite.client.InvokeXVMContract(fakeAddr, "bbb", nil, rpcx.Int32(1))
	suite.Require().Nil(err)
	suite.Require().True(result.Status == pb.Receipt_FAILED)
	suite.Require().Contains(string(result.Ret), "contract byte not correct")
}

//tc:调用合约，调用方法名为空，交易回执状态显示失败
func (suite *Snake) Test0407_InvokeContractEmptyMethod() {
	address := deployExampleContract(suite)

	result, err := suite.client.InvokeXVMContract(address, "", nil)
	suite.Require().Nil(err)
	suite.Require().True(result.Status == pb.Receipt_FAILED)
	suite.Require().Contains(string(result.Ret), "lack of method name")
}

//tc:部署合约，合约数据大小为10M以上，返回回执状态失败（待定）
func (suite *Snake) Test0408_Deploy10MContract() {
	// todo: wait for bitxhub to limit contract size
}

//tc:调用合约，调用参数不正确，交易回执状态显示失败
func (suite *Snake) Test0409_InvokeContractWrongArg() {
	address := deployExampleContract(suite)

	result, err := suite.client.InvokeXVMContract(address, "a", nil, rpcx.String("1"), rpcx.Int32(2))
	suite.Require().Nil(err)
	suite.Require().True(result.Status == pb.Receipt_FAILED)
	suite.Require().Contains(string(result.Ret), "not found allocate method")

	// incorrect function params
	result, err = suite.client.InvokeXVMContract(address, "a", nil, rpcx.Int32(1), rpcx.String("2"))
	suite.Require().Nil(err)
	suite.Require().True(result.Status == pb.Receipt_FAILED)

	result, err = suite.client.InvokeXVMContract(address, "a", nil, rpcx.String("1"), rpcx.String("2"))
	suite.Require().Nil(err)
	suite.Require().True(result.Status == pb.Receipt_FAILED)
}

//tc:调用合约，调用参数个数不正确，交易回执显示失败
func (suite *Snake) Test0410_InvokeContractWrongNumberArg() {
	address := deployExampleContract(suite)

	result, err := suite.client.InvokeXVMContract(address, "a", nil, rpcx.Int32(1), rpcx.Int32(2), rpcx.Int32(3))
	suite.Require().Nil(err)
	suite.Require().True(result.Status == pb.Receipt_FAILED)
}

func deployExampleContract(suite *Snake) *types.Address {
	contract, err := ioutil.ReadFile("testdata/example.wasm")
	suite.Require().Nil(err)

	address, err := suite.client.DeployContract(contract, nil)
	suite.Require().Nil(err)
	suite.Require().NotNil(address)
	return address
}
