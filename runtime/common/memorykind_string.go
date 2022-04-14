// Code generated by "stringer -type=MemoryKind -trimprefix=MemoryKind"; DO NOT EDIT.

package common

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[MemoryKindUnknown-0]
	_ = x[MemoryKindBool-1]
	_ = x[MemoryKindAddress-2]
	_ = x[MemoryKindString-3]
	_ = x[MemoryKindCharacter-4]
	_ = x[MemoryKindMetaType-5]
	_ = x[MemoryKindNumber-6]
	_ = x[MemoryKindArrayBase-7]
	_ = x[MemoryKindArrayLength-8]
	_ = x[MemoryKindDictionaryBase-9]
	_ = x[MemoryKindDictionarySize-10]
	_ = x[MemoryKindCompositeBase-11]
	_ = x[MemoryKindCompositeSize-12]
	_ = x[MemoryKindOptional-13]
	_ = x[MemoryKindNil-14]
	_ = x[MemoryKindVoid-15]
	_ = x[MemoryKindTypeValue-16]
	_ = x[MemoryKindPathValue-17]
	_ = x[MemoryKindCapabilityValue-18]
	_ = x[MemoryKindLinkValue-19]
	_ = x[MemoryKindStorageReferenceValue-20]
	_ = x[MemoryKindEphemeralReferenceValue-21]
	_ = x[MemoryKindInterpretedFunction-22]
	_ = x[MemoryKindHostFunction-23]
	_ = x[MemoryKindBoundFunction-24]
	_ = x[MemoryKindBigInt-25]
	_ = x[MemoryKindRawString-26]
	_ = x[MemoryKindAddressLocation-27]
	_ = x[MemoryKindBytes-28]
	_ = x[MemoryKindVariable-29]
	_ = x[MemoryKindValueToken-30]
	_ = x[MemoryKindSyntaxToken-31]
	_ = x[MemoryKindSpaceToken-32]
	_ = x[MemoryKindProgram-33]
	_ = x[MemoryKindIdentifier-34]
	_ = x[MemoryKindArgument-35]
	_ = x[MemoryKindBlock-36]
	_ = x[MemoryKindFunctionBlock-37]
	_ = x[MemoryKindParameter-38]
	_ = x[MemoryKindParameterList-39]
	_ = x[MemoryKindTransfer-40]
	_ = x[MemoryKindMembers-41]
	_ = x[MemoryKindTypeAnnotation-42]
	_ = x[MemoryKindDictionaryEntry-43]
	_ = x[MemoryKindFunctionDeclaration-44]
	_ = x[MemoryKindCompositeDeclaration-45]
	_ = x[MemoryKindInterfaceDeclaration-46]
	_ = x[MemoryKindEnumCaseDeclaration-47]
	_ = x[MemoryKindFieldDeclaration-48]
	_ = x[MemoryKindTransactionDeclaration-49]
	_ = x[MemoryKindImportDeclaration-50]
	_ = x[MemoryKindVariableDeclaration-51]
	_ = x[MemoryKindSpecialFunctionDeclaration-52]
	_ = x[MemoryKindPragmaDeclaration-53]
	_ = x[MemoryKindAssignmentStatement-54]
	_ = x[MemoryKindBreakStatement-55]
	_ = x[MemoryKindContinueStatement-56]
	_ = x[MemoryKindEmitStatement-57]
	_ = x[MemoryKindExpressionStatement-58]
	_ = x[MemoryKindForStatement-59]
	_ = x[MemoryKindIfStatement-60]
	_ = x[MemoryKindReturnStatement-61]
	_ = x[MemoryKindSwapStatement-62]
	_ = x[MemoryKindSwitchStatement-63]
	_ = x[MemoryKindWhileStatement-64]
	_ = x[MemoryKindBooleanExpression-65]
	_ = x[MemoryKindNilExpression-66]
	_ = x[MemoryKindStringExpression-67]
	_ = x[MemoryKindIntegerExpression-68]
	_ = x[MemoryKindFixedPointExpression-69]
	_ = x[MemoryKindArrayExpression-70]
	_ = x[MemoryKindDictionaryExpression-71]
	_ = x[MemoryKindIdentifierExpression-72]
	_ = x[MemoryKindInvocationExpression-73]
	_ = x[MemoryKindMemberExpression-74]
	_ = x[MemoryKindIndexExpression-75]
	_ = x[MemoryKindConditionalExpression-76]
	_ = x[MemoryKindUnaryExpression-77]
	_ = x[MemoryKindBinaryExpression-78]
	_ = x[MemoryKindFunctionExpression-79]
	_ = x[MemoryKindCastingExpression-80]
	_ = x[MemoryKindCreateExpression-81]
	_ = x[MemoryKindDestroyExpression-82]
	_ = x[MemoryKindReferenceExpression-83]
	_ = x[MemoryKindForceExpression-84]
	_ = x[MemoryKindPathExpression-85]
	_ = x[MemoryKindConstantSizedType-86]
	_ = x[MemoryKindDictionaryType-87]
	_ = x[MemoryKindFunctionType-88]
	_ = x[MemoryKindInstantiationType-89]
	_ = x[MemoryKindNominalType-90]
	_ = x[MemoryKindOptionalType-91]
	_ = x[MemoryKindReferenceType-92]
	_ = x[MemoryKindRestrictedType-93]
	_ = x[MemoryKindVariableSizedType-94]
	_ = x[MemoryKindPosition-95]
	_ = x[MemoryKindRange-96]
	_ = x[MemoryKindElaboration-97]
	_ = x[MemoryKindLast-98]
}

const _MemoryKind_name = "UnknownBoolAddressStringCharacterMetaTypeNumberArrayBaseArrayLengthDictionaryBaseDictionarySizeCompositeBaseCompositeSizeOptionalNilVoidTypeValuePathValueCapabilityValueLinkValueStorageReferenceValueEphemeralReferenceValueInterpretedFunctionHostFunctionBoundFunctionBigIntRawStringAddressLocationBytesVariableValueTokenSyntaxTokenSpaceTokenProgramIdentifierArgumentBlockFunctionBlockParameterParameterListTransferMembersTypeAnnotationDictionaryEntryFunctionDeclarationCompositeDeclarationInterfaceDeclarationEnumCaseDeclarationFieldDeclarationTransactionDeclarationImportDeclarationVariableDeclarationSpecialFunctionDeclarationPragmaDeclarationAssignmentStatementBreakStatementContinueStatementEmitStatementExpressionStatementForStatementIfStatementReturnStatementSwapStatementSwitchStatementWhileStatementBooleanExpressionNilExpressionStringExpressionIntegerExpressionFixedPointExpressionArrayExpressionDictionaryExpressionIdentifierExpressionInvocationExpressionMemberExpressionIndexExpressionConditionalExpressionUnaryExpressionBinaryExpressionFunctionExpressionCastingExpressionCreateExpressionDestroyExpressionReferenceExpressionForceExpressionPathExpressionConstantSizedTypeDictionaryTypeFunctionTypeInstantiationTypeNominalTypeOptionalTypeReferenceTypeRestrictedTypeVariableSizedTypePositionRangeElaborationLast"

var _MemoryKind_index = [...]uint16{0, 7, 11, 18, 24, 33, 41, 47, 56, 67, 81, 95, 108, 121, 129, 132, 136, 145, 154, 169, 178, 199, 222, 241, 253, 266, 272, 281, 296, 301, 309, 319, 330, 340, 347, 357, 365, 370, 383, 392, 405, 413, 420, 434, 449, 468, 488, 508, 527, 543, 565, 582, 601, 627, 644, 663, 677, 694, 707, 726, 738, 749, 764, 777, 792, 806, 823, 836, 852, 869, 889, 904, 924, 944, 964, 980, 995, 1016, 1031, 1047, 1065, 1082, 1098, 1115, 1134, 1149, 1163, 1180, 1194, 1206, 1223, 1234, 1246, 1259, 1273, 1290, 1298, 1303, 1314, 1318}

func (i MemoryKind) String() string {
	if i >= MemoryKind(len(_MemoryKind_index)-1) {
		return "MemoryKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _MemoryKind_name[_MemoryKind_index[i]:_MemoryKind_index[i+1]]
}
