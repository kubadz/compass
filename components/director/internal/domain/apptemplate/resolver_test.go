package apptemplate_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_ApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	appTemplateSys := fixModelAppTemplate(testID, testName)
	gqlAppTemplate := fixGQLAppTemplate(testID, testName)

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		ExpectedOutput    *graphql.ApplicationTemplate
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(appTemplateSys, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ToGraphQL", appTemplateSys).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Returns nil when application template not found",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, apperrors.NewNotFoundError("")).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(appTemplateSys, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(appTemplateSys, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ToGraphQL", appTemplateSys).Return(nil, testError).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()

			resolver := apptemplate.NewResolver(transact, appTemplateSvc, appTemplateConv)

			// WHEN
			result, err := resolver.ApplicationTemplate(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			appTemplateSvc.AssertExpectations(t)
			appTemplateConv.AssertExpectations(t)
		})
	}
}

func TestResolver_ApplicationTemplates(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)
	txGen := txtest.NewTransactionContextGenerator(testError)
	modelAppTemplates := []*model.ApplicationTemplate{
		fixModelAppTemplate("i1", "n1"),
		fixModelAppTemplate("i2", "n2"),
	}
	modelPage := fixModelAppTemplatePage(modelAppTemplates)
	gqlAppTemplates := []*graphql.ApplicationTemplate{
		fixGQLAppTemplate("i1", "n1"),
		fixGQLAppTemplate("i2", "n2"),
	}
	gqlPage := fixGQLAppTemplatePage(gqlAppTemplates)
	first := 2
	after := "test"
	gqlAfter := graphql.PageCursor(after)

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		ExpectedOutput    *graphql.ApplicationTemplatePage
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(modelPage, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("MultipleToGraphQL", modelAppTemplates).Return(gqlAppTemplates, nil).Once()
				return appTemplateConv
			},
			ExpectedOutput: &gqlPage,
		},
		{
			Name: "Returns error when getting application templates failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(model.ApplicationTemplatePage{}, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(modelPage, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert at least one of application templates to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(modelPage, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("MultipleToGraphQL", modelAppTemplates).Return(nil, testError).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()

			resolver := apptemplate.NewResolver(transact, appTemplateSvc, appTemplateConv)

			// WHEN
			result, err := resolver.ApplicationTemplates(ctx, &first, &gqlAfter)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			appTemplateSvc.AssertExpectations(t)
			appTemplateConv.AssertExpectations(t)
		})
	}
}

func TestResolver_CreateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	modelAppTemplate := fixModelAppTemplate(testID, testName)
	modelAppTemplateInput := fixModelAppTemplateInput(testName, fixApplicationCreateInputString())
	gqlAppTemplate := fixGQLAppTemplate(testID, testName)
	gqlAppTemplateInput := fixGQLAppTemplateInput(testName)

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		ExpectedOutput    *graphql.ApplicationTemplate
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Create", txtest.CtxWithDBMatcher(), *modelAppTemplateInput).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Returns error when can't convert input from graphql",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(model.ApplicationTemplateInput{}, testError).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when creating application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Create", txtest.CtxWithDBMatcher(), *modelAppTemplateInput).Return("", testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Create", txtest.CtxWithDBMatcher(), *modelAppTemplateInput).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Create", txtest.CtxWithDBMatcher(), *modelAppTemplateInput).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Create", txtest.CtxWithDBMatcher(), *modelAppTemplateInput).Return(modelAppTemplate.ID, nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(nil, testError).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()

			resolver := apptemplate.NewResolver(transact, appTemplateSvc, appTemplateConv)

			// WHEN
			result, err := resolver.CreateApplicationTemplate(ctx, *gqlAppTemplateInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			appTemplateSvc.AssertExpectations(t)
			appTemplateConv.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	modelAppTemplate := fixModelAppTemplate(testID, testName)
	modelAppTemplateInput := fixModelAppTemplateInput(testName, fixApplicationCreateInputString())
	gqlAppTemplate := fixGQLAppTemplate(testID, testName)
	gqlAppTemplateInput := fixGQLAppTemplateInput(testName)

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		ExpectedOutput    *graphql.ApplicationTemplate
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Returns error when can't convert input from graphql",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(model.ApplicationTemplateInput{}, testError).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when updating application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Update", txtest.CtxWithDBMatcher(), testID, *modelAppTemplateInput).Return(nil).Once()
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("InputFromGraphQL", *gqlAppTemplateInput).Return(*modelAppTemplateInput, nil).Once()
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(nil, testError).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()

			resolver := apptemplate.NewResolver(transact, appTemplateSvc, appTemplateConv)

			// WHEN
			result, err := resolver.UpdateApplicationTemplate(ctx, testID, *gqlAppTemplateInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			appTemplateSvc.AssertExpectations(t)
			appTemplateConv.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	txGen := txtest.NewTransactionContextGenerator(testError)

	modelAppTemplate := fixModelAppTemplate(testID, testName)
	gqlAppTemplate := fixGQLAppTemplate(testID, testName)

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppTemplateSvcFn  func() *automock.ApplicationTemplateService
		AppTemplateConvFn func() *automock.ApplicationTemplateConverter
		ExpectedOutput    *graphql.ApplicationTemplate
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(gqlAppTemplate, nil).Once()
				return appTemplateConv
			},
			ExpectedOutput: gqlAppTemplate,
		},
		{
			Name: "Returns error when getting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when deleting application template failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(testError).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				return appTemplateConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when can't convert application template to graphql",
			TxFn: txGen.ThatSucceeds,
			AppTemplateSvcFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return appTemplateSvc
			},
			AppTemplateConvFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ToGraphQL", modelAppTemplate).Return(nil, testError).Once()
				return appTemplateConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			appTemplateSvc := testCase.AppTemplateSvcFn()
			appTemplateConv := testCase.AppTemplateConvFn()
			resolver := apptemplate.NewResolver(transact, appTemplateSvc, appTemplateConv)

			// WHEN
			result, err := resolver.DeleteApplicationTemplate(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			appTemplateSvc.AssertExpectations(t)
			appTemplateConv.AssertExpectations(t)
		})
	}
}
