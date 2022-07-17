/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019-2022 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package stdlib

import (
	"fmt"

	"golang.org/x/crypto/sha3"

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/cadence/runtime/sema"
)

const authAccountFunctionDocString = `
Creates a new account, paid by the given existing account
`

var authAccountFunctionType = &sema.FunctionType{
	Parameters: []*sema.Parameter{
		{
			Identifier: "payer",
			TypeAnnotation: sema.NewTypeAnnotation(
				sema.AuthAccountType,
			),
		},
	},
	ReturnTypeAnnotation: sema.NewTypeAnnotation(
		sema.AuthAccountType,
	),
}

type EventEmitter interface {
	EmitEvent(
		inter *interpreter.Interpreter,
		eventType *sema.CompositeType,
		values []interpreter.Value,
		getLocationRange func() interpreter.LocationRange,
	)
}

type AuthAccountHandler interface {
	BalanceProvider
	AvailableBalanceProvider
	StorageUsedProvider
	StorageCapacityProvider
	AccountEncodedKeyAdditionHandler
	AccountEncodedKeyRevocationHandler
	AuthAccountKeysHandler
	AuthAccountContractsHandler
}

type AccountCreator interface {
	EventEmitter
	AuthAccountHandler
	// CreateAccount creates a new account.
	CreateAccount(payer common.Address) (address common.Address, err error)
}

func NewAuthAccountConstructor(creator AccountCreator) StandardLibraryValue {
	return NewStandardLibraryFunction(
		"AuthAccount",
		authAccountFunctionType,
		authAccountFunctionDocString,
		func(invocation interpreter.Invocation) interpreter.Value {

			payer, ok := invocation.Arguments[0].(interpreter.MemberAccessibleValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			inter := invocation.Interpreter
			getLocationRange := invocation.GetLocationRange

			inter.ExpectType(
				payer,
				sema.AuthAccountType,
				getLocationRange,
			)

			payerValue := payer.GetMember(
				inter,
				getLocationRange,
				sema.AuthAccountAddressField,
			)
			if payerValue == nil {
				panic(errors.NewUnexpectedError("payer address is not set"))
			}

			payerAddressValue, ok := payerValue.(interpreter.AddressValue)
			if !ok {
				panic(errors.NewUnexpectedError("payer address is not address"))
			}

			payerAddress := payerAddressValue.ToAddress()

			addressValue := interpreter.NewAddressValueFromConstructor(
				inter,
				func() (address common.Address) {
					var err error
					wrapPanic(func() {
						address, err = creator.CreateAccount(payerAddress)
					})
					if err != nil {
						panic(err)
					}

					return
				},
			)

			creator.EmitEvent(
				inter,
				AccountCreatedEventType,
				[]interpreter.Value{addressValue},
				getLocationRange,
			)

			return NewAuthAccountValue(
				inter,
				creator,
				addressValue,
			)
		},
	)
}

const getAuthAccountDocString = `
Returns the AuthAccount for the given address. Only available in scripts
`

var getAuthAccountFunctionType = &sema.FunctionType{
	Parameters: []*sema.Parameter{{
		Label:          sema.ArgumentLabelNotRequired,
		Identifier:     "address",
		TypeAnnotation: sema.NewTypeAnnotation(&sema.AddressType{}),
	}},
	ReturnTypeAnnotation: sema.NewTypeAnnotation(sema.AuthAccountType),
}

func NewGetAuthAccountFunction(handler AuthAccountHandler) StandardLibraryValue {
	return NewStandardLibraryFunction(
		"getAuthAccount",
		getAuthAccountFunctionType,
		getAuthAccountDocString,
		func(invocation interpreter.Invocation) interpreter.Value {
			accountAddress, ok := invocation.Arguments[0].(interpreter.AddressValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			gauge := invocation.Interpreter

			return NewAuthAccountValue(
				gauge,
				handler,
				accountAddress,
			)
		},
	)
}

func NewAuthAccountValue(
	gauge common.MemoryGauge,
	handler AuthAccountHandler,
	addressValue interpreter.AddressValue,
) interpreter.Value {
	return interpreter.NewAuthAccountValue(
		gauge,
		addressValue,
		newAccountBalanceGetFunction(gauge, handler, addressValue),
		newAccountAvailableBalanceGetFunction(gauge, handler, addressValue),
		newStorageUsedGetFunction(handler, addressValue),
		newStorageCapacityGetFunction(handler, addressValue),
		newAddPublicKeyFunction(gauge, handler, addressValue),
		newRemovePublicKeyFunction(gauge, handler, addressValue),
		func() interpreter.Value {
			return newAuthAccountContractsValue(
				gauge,
				handler,
				addressValue,
			)
		},
		func() interpreter.Value {
			return newAuthAccountKeysValue(
				gauge,
				handler,
				addressValue,
			)
		},
	)
}

type AuthAccountContractsHandler interface {
	AccountContractProvider
	AccountContractRemovalHandler
	AccountContractNamesProvider
}

func newAuthAccountContractsValue(
	gauge common.MemoryGauge,
	handler AuthAccountContractsHandler,
	addressValue interpreter.AddressValue,
) interpreter.Value {
	return interpreter.NewAuthAccountContractsValue(
		gauge,
		addressValue,
		newAuthAccountContractsChangeFunction(
			gauge,
			handler,
			addressValue,
			false,
		),
		newAuthAccountContractsChangeFunction(
			gauge,
			handler,
			addressValue,
			true,
		),
		newAccountContractsGetFunction(
			gauge,
			handler,
			addressValue,
		),
		newAuthAccountContractsRemoveFunction(
			gauge,
			handler,
			addressValue,
		),
		newAccountContractsGetNamesFunction(
			handler,
			addressValue,
		),
	)
}

type AuthAccountKeysHandler interface {
	AccountKeyProvider
	AccountKeyAdditionHandler
	AccountKeyRevocationHandler
}

func newAuthAccountKeysValue(
	gauge common.MemoryGauge,
	handler AuthAccountKeysHandler,
	addressValue interpreter.AddressValue,
) interpreter.Value {
	return interpreter.NewAuthAccountKeysValue(
		gauge,
		addressValue,
		newAccountKeysAddFunction(
			gauge,
			handler,
			addressValue,
		),
		newAccountKeysGetFunction(
			gauge,
			handler,
			addressValue,
		),
		newAccountKeysRevokeFunction(
			gauge,
			handler,
			addressValue,
		),
	)
}

type BalanceProvider interface {
	// GetAccountBalance gets accounts default flow token balance.
	GetAccountBalance(address common.Address) (uint64, error)
}

func newAccountBalanceGetFunction(
	gauge common.MemoryGauge,
	provider BalanceProvider,
	addressValue interpreter.AddressValue,
) func() interpreter.UFix64Value {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return func() interpreter.UFix64Value {
		return interpreter.NewUFix64Value(gauge, func() (balance uint64) {
			var err error
			wrapPanic(func() {
				balance, err = provider.GetAccountBalance(address)
			})
			if err != nil {
				panic(err)
			}

			return
		})
	}
}

type AvailableBalanceProvider interface {
	// GetAccountAvailableBalance gets accounts default flow token balance - balance that is reserved for storage.
	GetAccountAvailableBalance(address common.Address) (uint64, error)
}

func newAccountAvailableBalanceGetFunction(
	gauge common.MemoryGauge,
	provider AvailableBalanceProvider,
	addressValue interpreter.AddressValue,
) func() interpreter.UFix64Value {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return func() interpreter.UFix64Value {
		return interpreter.NewUFix64Value(gauge, func() (balance uint64) {
			var err error
			wrapPanic(func() {
				balance, err = provider.GetAccountAvailableBalance(address)
			})
			if err != nil {
				panic(err)
			}

			return
		})
	}
}

type StorageUsedProvider interface {
	CommitStorage(inter *interpreter.Interpreter, commitContractUpdates bool) error
	// GetStorageUsed gets storage used in bytes by the address at the moment of the function call.
	GetStorageUsed(address common.Address) (uint64, error)
}

func newStorageUsedGetFunction(
	provider StorageUsedProvider,
	addressValue interpreter.AddressValue,
) func(inter *interpreter.Interpreter) interpreter.UInt64Value {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return func(inter *interpreter.Interpreter) interpreter.UInt64Value {

		// NOTE: flush the cached values, so the host environment
		// can properly calculate the amount of storage used by the account
		const commitContractUpdates = false
		err := provider.CommitStorage(inter, commitContractUpdates)
		if err != nil {
			panic(err)
		}

		return interpreter.NewUInt64Value(
			inter,
			func() uint64 {
				var capacity uint64
				wrapPanic(func() {
					capacity, err = provider.GetStorageUsed(address)
				})
				if err != nil {
					panic(err)
				}
				return capacity
			},
		)
	}
}

type StorageCapacityProvider interface {
	CommitStorage(inter *interpreter.Interpreter, commitContractUpdates bool) error
	// GetStorageCapacity gets storage capacity in bytes on the address.
	GetStorageCapacity(address common.Address) (uint64, error)
}

func newStorageCapacityGetFunction(
	provider StorageCapacityProvider,
	addressValue interpreter.AddressValue,
) func(inter *interpreter.Interpreter) interpreter.UInt64Value {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return func(inter *interpreter.Interpreter) interpreter.UInt64Value {

		// NOTE: flush the cached values, so the host environment
		// can properly calculate the amount of storage available for the account
		const commitContractUpdates = false
		err := provider.CommitStorage(inter, commitContractUpdates)
		if err != nil {
			panic(err)
		}

		return interpreter.NewUInt64Value(
			inter,
			func() uint64 {
				var capacity uint64
				wrapPanic(func() {
					capacity, err = provider.GetStorageCapacity(address)
				})
				if err != nil {
					panic(err)
				}
				return capacity
			},
		)

	}
}

type AccountEncodedKeyAdditionHandler interface {
	EventEmitter
	// AddEncodedAccountKey appends an encoded key to an account.
	AddEncodedAccountKey(address common.Address, key []byte) error
}

func newAddPublicKeyFunction(
	gauge common.MemoryGauge,
	handler AccountEncodedKeyAdditionHandler,
	addressValue interpreter.AddressValue,
) *interpreter.HostFunctionValue {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return interpreter.NewHostFunctionValue(
		gauge,
		func(invocation interpreter.Invocation) interpreter.Value {
			publicKeyValue, ok := invocation.Arguments[0].(*interpreter.ArrayValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			publicKey, err := interpreter.ByteArrayValueToByteSlice(gauge, publicKeyValue)
			if err != nil {
				panic("addPublicKey requires the first argument to be a byte array")
			}

			wrapPanic(func() {
				err = handler.AddEncodedAccountKey(address, publicKey)
			})
			if err != nil {
				panic(err)
			}

			inter := invocation.Interpreter

			handler.EmitEvent(
				inter,
				AccountKeyAddedEventType,
				[]interpreter.Value{
					addressValue,
					publicKeyValue,
				},
				invocation.GetLocationRange,
			)

			return interpreter.VoidValue{}
		},
		sema.AuthAccountTypeAddPublicKeyFunctionType,
	)
}

type AccountEncodedKeyRevocationHandler interface {
	EventEmitter
	// RevokeEncodedAccountKey removes a key from an account by index, add returns the encoded key.
	RevokeEncodedAccountKey(address common.Address, index int) ([]byte, error)
}

func newRemovePublicKeyFunction(
	gauge common.MemoryGauge,
	handler AccountEncodedKeyRevocationHandler,
	addressValue interpreter.AddressValue,
) *interpreter.HostFunctionValue {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return interpreter.NewHostFunctionValue(
		gauge,
		func(invocation interpreter.Invocation) interpreter.Value {
			index, ok := invocation.Arguments[0].(interpreter.IntValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			var publicKey []byte
			var err error
			wrapPanic(func() {
				publicKey, err = handler.RevokeEncodedAccountKey(address, index.ToInt())
			})
			if err != nil {
				panic(err)
			}

			inter := invocation.Interpreter

			publicKeyValue := interpreter.ByteSliceToByteArrayValue(
				inter,
				publicKey,
			)

			handler.EmitEvent(
				inter,
				AccountKeyRemovedEventType,
				[]interpreter.Value{
					addressValue,
					publicKeyValue,
				},
				invocation.GetLocationRange,
			)

			return interpreter.VoidValue{}
		},
		sema.AuthAccountTypeRemovePublicKeyFunctionType,
	)
}

type AccountKeyAdditionHandler interface {
	EventEmitter
	// AddAccountKey appends a key to an account.
	AddAccountKey(address common.Address, key *PublicKey, algo sema.HashAlgorithm, weight int) (*AccountKey, error)
}

func newAccountKeysAddFunction(
	gauge common.MemoryGauge,
	handler AccountKeyAdditionHandler,
	addressValue interpreter.AddressValue,
) *interpreter.HostFunctionValue {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return interpreter.NewHostFunctionValue(
		gauge,
		func(invocation interpreter.Invocation) interpreter.Value {
			publicKeyValue, ok := invocation.Arguments[0].(*interpreter.CompositeValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			inter := invocation.Interpreter
			getLocationRange := invocation.GetLocationRange

			publicKey, err := NewPublicKeyFromValue(inter, getLocationRange, publicKeyValue)
			if err != nil {
				panic(err)
			}

			hashAlgo := NewHashAlgorithmFromValue(inter, getLocationRange, invocation.Arguments[1])
			weightValue, ok := invocation.Arguments[2].(interpreter.UFix64Value)
			if !ok {
				panic(errors.NewUnreachableError())
			}
			weight := weightValue.ToInt()

			var accountKey *AccountKey
			wrapPanic(func() {
				accountKey, err = handler.AddAccountKey(address, publicKey, hashAlgo, weight)
			})
			if err != nil {
				panic(err)
			}

			handler.EmitEvent(
				inter,
				AccountKeyAddedEventType,
				[]interpreter.Value{
					addressValue,
					publicKeyValue,
				},
				invocation.GetLocationRange,
			)

			return NewAccountKeyValue(
				inter,
				getLocationRange,
				accountKey,
				inter.PublicKeyValidationHandler,
			)
		},
		sema.AuthAccountKeysTypeAddFunctionType,
	)
}

type AccountKey struct {
	KeyIndex  int
	PublicKey *PublicKey
	HashAlgo  sema.HashAlgorithm
	Weight    int
	IsRevoked bool
}

type PublicKey struct {
	PublicKey []byte
	SignAlgo  sema.SignatureAlgorithm
}

type AccountKeyProvider interface {
	// GetAccountKey retrieves a key from an account by index.
	GetAccountKey(address common.Address, index int) (*AccountKey, error)
}

func newAccountKeysGetFunction(
	gauge common.MemoryGauge,
	provider AccountKeyProvider,
	addressValue interpreter.AddressValue,
) *interpreter.HostFunctionValue {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return interpreter.NewHostFunctionValue(
		gauge,
		func(invocation interpreter.Invocation) interpreter.Value {
			indexValue, ok := invocation.Arguments[0].(interpreter.IntValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}
			index := indexValue.ToInt()

			var err error
			var accountKey *AccountKey
			wrapPanic(func() {
				accountKey, err = provider.GetAccountKey(address, index)
			})

			if err != nil {
				panic(err)
			}

			// Here it is expected the host function to return a nil key, if a key is not found at the given index.
			// This is done because, if the host function returns an error when a key is not found, then
			// currently there's no way to distinguish between a 'key not found error' vs other internal errors.
			if accountKey == nil {
				return interpreter.NewNilValue(invocation.Interpreter)
			}

			inter := invocation.Interpreter

			return interpreter.NewSomeValueNonCopying(
				inter,
				NewAccountKeyValue(
					inter,
					invocation.GetLocationRange,
					accountKey,
					// public keys are assumed to be already validated.
					func(
						_ *interpreter.Interpreter,
						_ func() interpreter.LocationRange,
						_ *interpreter.CompositeValue,
					) error {
						return nil
					},
				),
			)
		},
		sema.AccountKeysTypeGetFunctionType,
	)
}

type AccountKeyRevocationHandler interface {
	EventEmitter
	// RevokeAccountKey removes a key from an account by index.
	RevokeAccountKey(address common.Address, index int) (*AccountKey, error)
}

func newAccountKeysRevokeFunction(
	gauge common.MemoryGauge,
	handler AccountKeyRevocationHandler,
	addressValue interpreter.AddressValue,
) *interpreter.HostFunctionValue {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return interpreter.NewHostFunctionValue(
		gauge,
		func(invocation interpreter.Invocation) interpreter.Value {
			indexValue, ok := invocation.Arguments[0].(interpreter.IntValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}
			index := indexValue.ToInt()

			var err error
			var accountKey *AccountKey
			wrapPanic(func() {
				accountKey, err = handler.RevokeAccountKey(address, index)
			})
			if err != nil {
				panic(err)
			}

			// Here it is expected the host function to return a nil key, if a key is not found at the given index.
			// This is done because, if the host function returns an error when a key is not found, then
			// currently there's no way to distinguish between a 'key not found error' vs other internal errors.
			if accountKey == nil {
				return interpreter.NewNilValue(invocation.Interpreter)
			}

			inter := invocation.Interpreter

			handler.EmitEvent(
				inter,
				AccountKeyRemovedEventType,
				[]interpreter.Value{
					addressValue,
					indexValue,
				},
				invocation.GetLocationRange,
			)

			return interpreter.NewSomeValueNonCopying(
				inter,
				NewAccountKeyValue(
					inter,
					invocation.GetLocationRange,
					accountKey,
					// public keys are assumed to be already validated.
					func(
						_ *interpreter.Interpreter,
						_ func() interpreter.LocationRange,
						_ *interpreter.CompositeValue,
					) error {
						return nil
					},
				),
			)
		},
		sema.AuthAccountKeysTypeRevokeFunctionType,
	)
}

type PublicAccountKeysHandler interface {
	AccountKeyProvider
}

func newPublicAccountKeysValue(
	gauge common.MemoryGauge,
	handler PublicAccountKeysHandler,
	addressValue interpreter.AddressValue,
) interpreter.Value {
	return interpreter.NewPublicAccountKeysValue(
		gauge,
		addressValue,
		newAccountKeysGetFunction(
			gauge,
			handler,
			addressValue,
		),
	)
}

type PublicAccountContractsHandler interface {
	AccountContractNamesProvider
	AccountContractProvider
}

func newPublicAccountContractsValue(
	gauge common.MemoryGauge,
	handler PublicAccountContractsHandler,
	addressValue interpreter.AddressValue,
) interpreter.Value {
	return interpreter.NewPublicAccountContractsValue(
		gauge,
		addressValue,
		newAccountContractsGetFunction(
			gauge,
			handler,
			addressValue,
		),
		newAccountContractsGetNamesFunction(
			handler,
			addressValue,
		),
	)
}

type AccountContractNamesProvider interface {
	// GetAccountContractNames returns the names of all contracts deployed in an account.
	GetAccountContractNames(address common.Address) ([]string, error)
}

func newAccountContractsGetNamesFunction(
	provider AccountContractNamesProvider,
	addressValue interpreter.AddressValue,
) func(
	inter *interpreter.Interpreter,
	getLocationRange func() interpreter.LocationRange,
) *interpreter.ArrayValue {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return func(
		inter *interpreter.Interpreter,
		getLocationRange func() interpreter.LocationRange,
	) *interpreter.ArrayValue {
		var names []string
		var err error
		wrapPanic(func() {
			names, err = provider.GetAccountContractNames(address)
		})
		if err != nil {
			panic(err)
		}

		values := make([]interpreter.Value, len(names))
		for i, name := range names {
			memoryUsage := common.NewStringMemoryUsage(len(name))
			values[i] = interpreter.NewStringValue(
				inter,
				memoryUsage,
				func() string {
					return name
				},
			)
		}

		arrayType := interpreter.NewVariableSizedStaticType(
			inter,
			interpreter.NewPrimitiveStaticType(
				inter,
				interpreter.PrimitiveStaticTypeString,
			),
		)

		return interpreter.NewArrayValue(
			inter,
			getLocationRange,
			arrayType,
			common.Address{},
			values...,
		)
	}
}

type AccountContractProvider interface {
	// GetAccountContractCode returns the code associated with an account contract.
	GetAccountContractCode(address common.Address, name string) ([]byte, error)
}

func newAccountContractsGetFunction(
	gauge common.MemoryGauge,
	provider AccountContractProvider,
	addressValue interpreter.AddressValue,
) *interpreter.HostFunctionValue {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return interpreter.NewHostFunctionValue(
		gauge,
		func(invocation interpreter.Invocation) interpreter.Value {
			nameValue, ok := invocation.Arguments[0].(*interpreter.StringValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}
			name := nameValue.Str

			var code []byte
			var err error
			wrapPanic(func() {
				code, err = provider.GetAccountContractCode(address, name)
			})
			if err != nil {
				panic(err)
			}

			if len(code) > 0 {
				return interpreter.NewSomeValueNonCopying(
					invocation.Interpreter,
					interpreter.NewDeployedContractValue(
						invocation.Interpreter,
						addressValue,
						nameValue,
						interpreter.ByteSliceToByteArrayValue(
							invocation.Interpreter,
							code,
						),
					),
				)
			} else {
				return interpreter.NewNilValue(invocation.Interpreter)
			}
		},
		sema.AuthAccountContractsTypeGetFunctionType,
	)
}

type AccountContractAdditionHandler interface {
	EventEmitter
	AccountContractProvider
}

// newAuthAccountContractsChangeFunction called when e.g.
// - adding: `AuthAccount.contracts.add(name: "Foo", code: [...])` (isUpdate = false)
// - updating: `AuthAccount.contracts.update__experimental(name: "Foo", code: [...])` (isUpdate = true)
//
func newAuthAccountContractsChangeFunction(
	gauge common.MemoryGauge,
	handler AccountContractAdditionHandler,
	addressValue interpreter.AddressValue,
	isUpdate bool,
) *interpreter.HostFunctionValue {
	return interpreter.NewHostFunctionValue(
		gauge,
		func(invocation interpreter.Invocation) interpreter.Value {

			const requiredArgumentCount = 2

			nameValue, ok := invocation.Arguments[0].(*interpreter.StringValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			newCodeValue, ok := invocation.Arguments[1].(*interpreter.ArrayValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			// TODO:
			//constructorArguments := invocation.Arguments[requiredArgumentCount:]
			//constructorArgumentTypes := invocation.ArgumentTypes[requiredArgumentCount:]

			code, err := interpreter.ByteArrayValueToByteSlice(gauge, newCodeValue)
			if err != nil {
				panic(errors.NewDefaultUserError("add requires the second argument to be an array"))
			}

			// Get the existing code

			nameArgument := nameValue.Str

			if nameArgument == "" {
				panic(errors.NewDefaultUserError(
					"contract name argument cannot be empty." +
						"it must match the name of the deployed contract declaration or contract interface declaration",
				))
			}

			address := addressValue.ToAddress()
			existingCode, err := handler.GetAccountContractCode(address, nameArgument)
			if err != nil {
				panic(err)
			}

			if isUpdate {
				// We are updating an existing contract.
				// Ensure that there's a contract/contract-interface with the given name exists already

				if len(existingCode) == 0 {
					panic(errors.NewDefaultUserError(
						"cannot update non-existing contract with name %q in account %s",
						nameArgument,
						address.ShortHexWithPrefix(),
					))
				}

			} else {
				// We are adding a new contract.
				// Ensure that no contract/contract interface with the given name exists already

				if len(existingCode) > 0 {
					panic(errors.NewDefaultUserError(
						"cannot overwrite existing contract with name %q in account %s",
						nameArgument,
						address.ShortHexWithPrefix(),
					))
				}
			}

			// TODO:
			//// Check the code
			//
			//location := common.NewAddressLocation(invocation.Interpreter, address, nameArgument)
			//
			//context := startContext.WithLocation(location)
			//
			//handleContractUpdateError := func(err error) {
			//	if err == nil {
			//		return
			//	}
			//
			//	// Update the code for the error pretty printing
			//	// NOTE: only do this when an error occurs
			//
			//	context.SetCode(context.Location, code)
			//
			//	panic(&InvalidContractDeploymentError{
			//		Err:           err,
			//		LocationRange: invocation.GetLocationRange(),
			//	})
			//}
			//
			//// NOTE: do NOT use the program obtained from the host environment, as the current program.
			//// Always re-parse and re-check the new program.
			//
			//// NOTE: *DO NOT* store the program – the new or updated program
			//// should not be effective during the execution
			//
			//const storeProgram = false
			//
			//program, err := r.parseAndCheckProgram(
			//	code,
			//	context,
			//	storeProgram,
			//	importResolutionResults{},
			//)
			//if err != nil {
			//	// Update the code for the error pretty printing
			//	// NOTE: only do this when an error occurs
			//
			//	context.SetCode(context.Location, code)
			//
			//	panic(&InvalidContractDeploymentError{
			//		Err:           err,
			//		LocationRange: invocation.GetLocationRange(),
			//	})
			//}
			//
			//// The code may declare exactly one contract or one contract interface.
			//
			//var contractTypes []*sema.CompositeType
			//var contractInterfaceTypes []*sema.InterfaceType
			//
			//program.Elaboration.GlobalTypes.Foreach(func(_ string, variable *sema.Variable) {
			//	switch ty := variable.Type.(type) {
			//	case *sema.CompositeType:
			//		if ty.Kind == common.CompositeKindContract {
			//			contractTypes = append(contractTypes, ty)
			//		}
			//
			//	case *sema.InterfaceType:
			//		if ty.CompositeKind == common.CompositeKindContract {
			//			contractInterfaceTypes = append(contractInterfaceTypes, ty)
			//		}
			//	}
			//})
			//
			//var deployedType sema.Type
			//var contractType *sema.CompositeType
			//var contractInterfaceType *sema.InterfaceType
			//var declaredName string
			//var declarationKind common.DeclarationKind
			//
			//switch {
			//case len(contractTypes) == 1 && len(contractInterfaceTypes) == 0:
			//	contractType = contractTypes[0]
			//	declaredName = contractType.Identifier
			//	deployedType = contractType
			//	declarationKind = common.DeclarationKindContract
			//case len(contractInterfaceTypes) == 1 && len(contractTypes) == 0:
			//	contractInterfaceType = contractInterfaceTypes[0]
			//	declaredName = contractInterfaceType.Identifier
			//	deployedType = contractInterfaceType
			//	declarationKind = common.DeclarationKindContractInterface
			//}
			//
			//if deployedType == nil {
			//	// Update the code for the error pretty printing
			//	// NOTE: only do this when an error occurs
			//
			//	context.SetCode(context.Location, code)
			//
			//	panic(errors.NewDefaultUserError(
			//		"invalid %s: the code must declare exactly one contract or contract interface",
			//		declarationKind.Name(),
			//	))
			//}
			//
			//// The declared contract or contract interface must have the name
			//// passed to the constructor as the first argument
			//
			//if declaredName != nameArgument {
			//	// Update the code for the error pretty printing
			//	// NOTE: only do this when an error occurs
			//
			//	context.SetCode(context.Location, code)
			//
			//	panic(errors.NewDefaultUserError(
			//		"invalid %s: the name argument must match the name of the declaration: got %q, expected %q",
			//		declarationKind.Name(),
			//		nameArgument,
			//		declaredName,
			//	))
			//}
			//
			//// Validate the contract update (if enabled)
			//
			//if isUpdate {
			//
			//	oldCode, err := r.getCode(context)
			//	handleContractUpdateError(err)
			//
			//	oldProgram, err := parser.ParseProgram(string(oldCode), gauge)
			//
			//	if !ignoreUpdatedProgramParserError(err) {
			//		handleContractUpdateError(err)
			//	}
			//
			//	validator := NewContractUpdateValidator(
			//		context.Location,
			//		nameArgument,
			//		oldProgram,
			//		program.Program,
			//	)
			//	err = validator.Validate()
			//	handleContractUpdateError(err)
			//}

			inter := invocation.Interpreter

			// TODO:
			//err = r.updateAccountContractCode(
			//	program,
			//	context,
			//	storage,
			//	declaredName,
			//	code,
			//	addressValue,
			//	contractType,
			//	constructorArguments,
			//	constructorArgumentTypes,
			//	updateAccountContractCodeOptions{
			//		createContract: !isUpdate,
			//	},
			//)
			//if err != nil {
			//	// Update the code for the error pretty printing
			//	// NOTE: only do this when an error occurs
			//
			//	context.SetCode(context.Location, code)
			//
			//	panic(err)
			//}

			var eventType *sema.CompositeType

			if isUpdate {
				eventType = AccountContractUpdatedEventType
			} else {
				eventType = AccountContractAddedEventType
			}

			codeHashValue := CodeToHashValue(inter, code)

			handler.EmitEvent(
				inter,
				eventType,
				[]interpreter.Value{
					addressValue,
					codeHashValue,
					nameValue,
				},
				invocation.GetLocationRange,
			)

			return interpreter.NewDeployedContractValue(
				inter,
				addressValue,
				nameValue,
				newCodeValue,
			)
		},
		sema.AuthAccountContractsTypeAddFunctionType,
	)
}

//
//func (r *interpreterRuntime) instantiateContract(
//	program *interpreter.Program,
//	context Context,
//	address common.Address,
//	contractType *sema.CompositeType,
//	constructorArguments []interpreter.Value,
//	argumentTypes []sema.Type,
//	storage *Storage,
//	environment *Environment,
//) (
//	*interpreter.CompositeValue,
//	error,
//) {
//	parameterTypes := make([]sema.Type, len(contractType.ConstructorParameters))
//
//	for i, constructorParameter := range contractType.ConstructorParameters {
//		parameterTypes[i] = constructorParameter.TypeAnnotation.Type
//	}
//
//	// Check argument count
//
//	argumentCount := len(argumentTypes)
//	parameterCount := len(parameterTypes)
//
//	if argumentCount < parameterCount {
//		return nil, errors.NewDefaultUserError(
//			"invalid argument count, too few arguments: expected %d, got %d, next missing argument: `%s`",
//			parameterCount, argumentCount,
//			parameterTypes[argumentCount],
//		)
//	} else if argumentCount > parameterCount {
//		return nil, errors.NewDefaultUserError(
//			"invalid argument count, too many arguments: expected %d, got %d",
//			parameterCount,
//			argumentCount,
//		)
//	}
//
//	// argumentCount now equals to parameterCount
//
//	// Check arguments match parameter
//
//	for i := 0; i < argumentCount; i++ {
//		argumentType := argumentTypes[i]
//		parameterTye := parameterTypes[i]
//		if !sema.IsSubType(argumentType, parameterTye) {
//			return nil, errors.NewDefaultUserError(
//				"invalid argument %d: expected type `%s`, got `%s`",
//				i,
//				parameterTye,
//				argumentType,
//			)
//		}
//	}
//
//	// Use a custom contract value handler that detects if the requested contract value
//	// is for the contract declaration that is being deployed.
//	//
//	// If the contract is the deployed contract, instantiate it using
//	// the provided constructor and given arguments.
//	//
//	// If the contract is not the deployed contract, load it from storage.
//
//	var contract *interpreter.CompositeValue
//
//	// TODO:
//	//interpreter.WithContractValueHandler(
//	//	func(
//	//		inter *interpreter.Interpreter,
//	//		compositeType *sema.CompositeType,
//	//		constructorGenerator func(common.Address) *interpreter.HostFunctionValue,
//	//		invocationRange ast.Range,
//	//	) *interpreter.CompositeValue {
//	//
//	//		constructor := constructorGenerator(address)
//	//
//	//		// If the contract is the deployed contract, instantiate it using
//	//		// the provided constructor and given arguments
//	//
//	//		if compositeType.Location == contractType.Location &&
//	//			compositeType.Identifier == contractType.Identifier {
//	//
//	//			value, err := inter.InvokeFunctionValue(
//	//				constructor,
//	//				constructorArguments,
//	//				argumentTypes,
//	//				parameterTypes,
//	//				invocationRange,
//	//			)
//	//			if err != nil {
//	//				panic(err)
//	//			}
//	//
//	//			return value.(*interpreter.CompositeValue)
//	//		}
//	//
//	//		// The contract is not the deployed contract, load it from storage
//	//		return r.loadContract(
//	//			inter,
//	//			compositeType,
//	//			constructorGenerator,
//	//			invocationRange,
//	//			storage,
//	//		)
//	//	},
//	//
//
//	_, inter, err := r.interpret(
//		program,
//		context,
//		storage,
//		environment,
//		nil,
//	)
//
//	if err != nil {
//		return nil, err
//	}
//
//	variable, ok := inter.Globals.Get(contractType.Identifier)
//	if !ok {
//		return nil, errors.NewDefaultUserError(
//			"cannot find contract: `%s`",
//			contractType.Identifier,
//		)
//	}
//
//	contract = variable.GetValue().(*interpreter.CompositeValue)
//
//	return contract, err
//}
//
//// ignoreUpdatedProgramParserError determines if the parsing error
//// for a program that is being updated can be ignored.
//func ignoreUpdatedProgramParserError(err error) bool {
//	parserError, ok := err.(parser.Error)
//	if !ok {
//		return false
//	}
//
//	// Are all parse errors ones that can be ignored?
//	for _, parseError := range parserError.Errors {
//		// Missing commas in parameter lists were reported starting
//		// with https://github.com/onflow/cadence/pull/1073.
//		// Allow existing contracts with such an error to be updated
//		_, ok := parseError.(*parser.MissingCommaInParameterListError)
//		if !ok {
//			return false
//		}
//	}
//
//	return true
//}
//
//type updateAccountContractCodeOptions struct {
//	createContract bool
//}
//
//// updateAccountContractCode updates an account contract's code.
//// This function is only used for the new account code/contract API.
////
//func (r *interpreterRuntime) updateAccountContractCode(
//	program *interpreter.Program,
//	context Context,
//	storage *Storage,
//	name string,
//	code []byte,
//	addressValue interpreter.AddressValue,
//	contractType *sema.CompositeType,
//	constructorArguments []interpreter.Value,
//	constructorArgumentTypes []sema.Type,
//	environment *Environment,
//	options updateAccountContractCodeOptions,
//) error {
//	// If the code declares a contract, instantiate it and store it.
//	//
//	// This function might be called when
//	// 1. A contract is deployed (contractType is non-nil).
//	// 2. A contract interface is deployed (contractType is nil).
//	//
//	// If a contract is deployed, it is only instantiated
//	// when options.createContract is true,
//	// i.e. the Cadence `add` function is used.
//	// If the Cadence `update__experimental` function is used,
//	// the new contract will NOT be deployed (options.createContract is false).
//
//	var contractValue *interpreter.CompositeValue
//
//	createContract := contractType != nil && options.createContract
//
//	address := addressValue.ToAddress()
//
//	var err error
//
//	if createContract {
//		contractValue, err = r.instantiateContract(
//			program,
//			context,
//			address,
//			contractType,
//			constructorArguments,
//			constructorArgumentTypes,
//			storage,
//			environment,
//		)
//
//		if err != nil {
//			return err
//		}
//	}
//
//	// NOTE: only update account code if contract instantiation succeeded
//	wrapPanic(func() {
//		err = context.Interface.UpdateAccountContractCode(address, name, code)
//	})
//	if err != nil {
//		return err
//	}
//
//	if createContract {
//		// NOTE: the contract recording delays the write
//		// until the end of the execution of the program
//
//		storage.recordContractUpdate(
//			addressValue.ToAddress(),
//			name,
//			contractValue,
//		)
//	}
//
//	return nil
//}

type AccountContractRemovalHandler interface {
	EventEmitter
	AccountContractProvider
	RemoveAccountContractCode(address common.Address, name string) error
	RecordContractRemoval(address common.Address, name string)
}

func newAuthAccountContractsRemoveFunction(
	gauge common.MemoryGauge,
	handler AccountContractRemovalHandler,
	addressValue interpreter.AddressValue,
) *interpreter.HostFunctionValue {

	// Converted addresses can be cached and don't have to be recomputed on each function invocation
	address := addressValue.ToAddress()

	return interpreter.NewHostFunctionValue(
		gauge,
		func(invocation interpreter.Invocation) interpreter.Value {

			inter := invocation.Interpreter
			nameValue, ok := invocation.Arguments[0].(*interpreter.StringValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}
			name := nameValue.Str

			// Get the current code

			var code []byte
			var err error
			wrapPanic(func() {
				code, err = handler.GetAccountContractCode(address, name)
			})
			if err != nil {
				panic(err)
			}

			// Only remove the contract code, remove the contract value, and emit an event,
			// if there is currently code deployed for the given contract name

			if len(code) > 0 {

				// NOTE: *DO NOT* call SetProgram – the program removal
				// should not be effective during the execution, only after

				existingProgram, err := parser.ParseProgram(string(code), gauge)

				// If the existing code is not parsable (i.e: `err != nil`),
				// that shouldn't be a reason to fail the contract removal.
				// Therefore, validate only if the code is a valid one.
				if err == nil && containsEnumsInProgram(existingProgram) {
					panic(&ContractRemovalError{
						Name:          name,
						LocationRange: invocation.GetLocationRange(),
					})
				}

				wrapPanic(func() {
					err = handler.RemoveAccountContractCode(address, name)
				})
				if err != nil {
					panic(err)
				}

				// NOTE: the contract recording function delays the write
				// until the end of the execution of the program

				handler.RecordContractRemoval(address, name)

				codeHashValue := CodeToHashValue(inter, code)

				handler.EmitEvent(
					inter,
					AccountContractRemovedEventType,
					[]interpreter.Value{
						addressValue,
						codeHashValue,
						nameValue,
					},
					invocation.GetLocationRange,
				)

				return interpreter.NewSomeValueNonCopying(
					inter,
					interpreter.NewDeployedContractValue(
						inter,
						addressValue,
						nameValue,
						interpreter.ByteSliceToByteArrayValue(
							inter,
							code,
						),
					),
				)
			} else {
				return interpreter.NewNilValue(invocation.Interpreter)
			}
		},
		sema.AuthAccountContractsTypeRemoveFunctionType,
	)
}

// ContractRemovalError
//
type ContractRemovalError struct {
	Name string
	interpreter.LocationRange
}

var _ errors.UserError = &ContractRemovalError{}

func (*ContractRemovalError) IsUserError() {}

func (e *ContractRemovalError) Error() string {
	return fmt.Sprintf("cannot remove contract `%s`", e.Name)
}

const getAccountFunctionDocString = `
Returns the public account for the given address
`

var getAccountFunctionType = &sema.FunctionType{
	Parameters: []*sema.Parameter{
		{
			Label:      sema.ArgumentLabelNotRequired,
			Identifier: "address",
			TypeAnnotation: sema.NewTypeAnnotation(
				&sema.AddressType{},
			),
		},
	},
	ReturnTypeAnnotation: sema.NewTypeAnnotation(
		sema.PublicAccountType,
	),
}

type PublicAccountHandler interface {
	BalanceProvider
	AvailableBalanceProvider
	StorageUsedProvider
	StorageCapacityProvider
	PublicAccountKeysHandler
	PublicAccountContractsHandler
}

func NewGetAccountFunction(handler PublicAccountHandler) StandardLibraryValue {
	return NewStandardLibraryFunction(
		"getAccount",
		getAccountFunctionType,
		getAccountFunctionDocString,
		func(invocation interpreter.Invocation) interpreter.Value {
			accountAddress, ok := invocation.Arguments[0].(interpreter.AddressValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			return NewPublicAccountValue(
				invocation.Interpreter,
				handler,
				accountAddress,
			)
		},
	)
}

func NewPublicAccountValue(
	gauge common.MemoryGauge,
	handler PublicAccountHandler,
	addressValue interpreter.AddressValue,
) interpreter.Value {
	return interpreter.NewPublicAccountValue(
		gauge,
		addressValue,
		newAccountBalanceGetFunction(gauge, handler, addressValue),
		newAccountAvailableBalanceGetFunction(gauge, handler, addressValue),
		newStorageUsedGetFunction(handler, addressValue),
		newStorageCapacityGetFunction(handler, addressValue),
		func() interpreter.Value {
			return newPublicAccountKeysValue(gauge, handler, addressValue)
		},
		func() interpreter.Value {
			return newPublicAccountContractsValue(gauge, handler, addressValue)
		},
	)
}

func NewAccountKeyValue(
	inter *interpreter.Interpreter,
	getLocationRange func() interpreter.LocationRange,
	accountKey *AccountKey,
	validatePublicKey interpreter.PublicKeyValidationHandlerFunc,
) interpreter.Value {
	return interpreter.NewAccountKeyValue(
		inter,
		interpreter.NewIntValueFromInt64(inter, int64(accountKey.KeyIndex)),
		NewPublicKeyValue(
			inter,
			getLocationRange,
			accountKey.PublicKey,
			validatePublicKey,
		),
		NewHashAlgorithmCase(
			interpreter.UInt8Value(accountKey.HashAlgo.RawValue()),
		),
		interpreter.NewUFix64ValueWithInteger(
			inter, func() uint64 {
				return uint64(accountKey.Weight)
			},
		),
		interpreter.BoolValue(accountKey.IsRevoked),
	)
}

func NewPublicKeyFromValue(
	inter *interpreter.Interpreter,
	getLocationRange func() interpreter.LocationRange,
	publicKey interpreter.MemberAccessibleValue,
) (
	*PublicKey,
	error,
) {
	// publicKey field
	key := publicKey.GetMember(inter, getLocationRange, sema.PublicKeyPublicKeyField)

	byteArray, err := interpreter.ByteArrayValueToByteSlice(inter, key)
	if err != nil {
		return nil, errors.NewUnexpectedError("public key needs to be a byte array. %w", err)
	}

	// sign algo field
	signAlgoField := publicKey.GetMember(inter, getLocationRange, sema.PublicKeySignAlgoField)
	if signAlgoField == nil {
		return nil, errors.NewUnexpectedError("sign algorithm is not set")
	}

	signAlgoValue, ok := signAlgoField.(*interpreter.SimpleCompositeValue)
	if !ok {
		return nil, errors.NewUnexpectedError(
			"sign algorithm does not belong to type: %s",
			sema.SignatureAlgorithmType.QualifiedString(),
		)
	}

	rawValue := signAlgoValue.GetMember(inter, getLocationRange, sema.EnumRawValueFieldName)
	if rawValue == nil {
		return nil, errors.NewDefaultUserError("sign algorithm raw value is not set")
	}

	signAlgoRawValue, ok := rawValue.(interpreter.UInt8Value)
	if !ok {
		return nil, errors.NewUnexpectedError(
			"sign algorithm raw-value does not belong to type: %s",
			sema.UInt8Type.QualifiedString(),
		)
	}

	return &PublicKey{
		PublicKey: byteArray,
		SignAlgo:  sema.SignatureAlgorithm(signAlgoRawValue.ToInt()),
	}, nil
}

func NewPublicKeyValue(
	inter *interpreter.Interpreter,
	getLocationRange func() interpreter.LocationRange,
	publicKey *PublicKey,
	validatePublicKey interpreter.PublicKeyValidationHandlerFunc,
) *interpreter.CompositeValue {
	return interpreter.NewPublicKeyValue(
		inter,
		getLocationRange,
		interpreter.ByteSliceToByteArrayValue(
			inter,
			publicKey.PublicKey,
		),
		NewSignatureAlgorithmCase(
			interpreter.UInt8Value(publicKey.SignAlgo.RawValue()),
		),
		func(
			inter *interpreter.Interpreter,
			getLocationRange func() interpreter.LocationRange,
			publicKeyValue *interpreter.CompositeValue,
		) error {
			return validatePublicKey(inter, getLocationRange, publicKeyValue)
		},
	)
}

func NewHashAlgorithmFromValue(
	inter *interpreter.Interpreter,
	getLocationRange func() interpreter.LocationRange,
	value interpreter.Value,
) sema.HashAlgorithm {
	hashAlgoValue := value.(*interpreter.SimpleCompositeValue)

	rawValue := hashAlgoValue.GetMember(inter, getLocationRange, sema.EnumRawValueFieldName)
	if rawValue == nil {
		panic("cannot find hash algorithm raw value")
	}

	hashAlgoRawValue := rawValue.(interpreter.UInt8Value)

	return sema.HashAlgorithm(hashAlgoRawValue.ToInt())
}

func CodeToHashValue(inter *interpreter.Interpreter, code []byte) *interpreter.ArrayValue {
	codeHash := sha3.Sum256(code)
	return interpreter.ByteSliceToByteArrayValue(inter, codeHash[:])
}