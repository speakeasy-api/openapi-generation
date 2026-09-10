package ast

import (
	"slices"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPublicExports(t *testing.T) {
	t.Parallel()

	t.Run("builds tree from rendered typedef", func(t *testing.T) {
		t.Parallel()

		model := publicExportTestType("CreateChatCompletionResponse")
		model.Extensions = &TypeDefExtensions{
			PublicExports: []extensions.PublicExport{
				{Group: "chat.completions", Name: "chat_completion"},
			},
		}

		a := publicExportTestAST(model, model)
		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.RootChildren, 1)
		assert.Equal(t, "chat", result.RootChildren[0].Name)
		require.Len(t, result.Groups, 2)

		chatGroup := result.Groups[0]
		assert.Equal(t, "chat", chatGroup.Group)
		require.Len(t, chatGroup.Children, 1)
		assert.Equal(t, "completions", chatGroup.Children[0].Name)
		assert.Empty(t, chatGroup.Exports)

		completionsGroup := result.Groups[1]
		assert.Equal(t, "chat.completions", completionsGroup.Group)
		require.Len(t, completionsGroup.Exports, 1)
		assert.Equal(t, "chat_completion", completionsGroup.Exports[0].Name)
		assert.Same(t, model, completionsGroup.Exports[0].Target)
	})

	t.Run("resolves nullable wrapper to rendered concrete child", func(t *testing.T) {
		t.Parallel()

		audio := publicExportTestType("ChatCompletionResponseMessageAudio")
		wrapper := &TypeDef{
			Name: "ChatCompletionResponseMessageAudio",
			Type: DataTypeUnion,
			AssociatedTypes: TypeDefs{
				audio,
				{Type: DataTypeAny, ContainsNull: true},
			},
			Extensions: &TypeDefExtensions{
				PublicExports: []extensions.PublicExport{
					{Group: "chat.completions", Name: "chat_completion_audio"},
				},
			},
		}

		a := publicExportTestAST(wrapper, audio)
		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.Groups, 2)
		completionsGroup := result.Groups[1]
		require.Len(t, completionsGroup.Exports, 1)
		assert.Equal(t, "chat_completion_audio", completionsGroup.Exports[0].Name)
		assert.Same(t, audio, completionsGroup.Exports[0].Target)
	})

	t.Run("resolves wrapper field to rendered nested child", func(t *testing.T) {
		t.Parallel()

		audio := publicExportTestType("Audio")
		wrapper := &TypeDef{
			Name: "ChatCompletionResponseMessage",
			Type: DataTypeClass,
			Fields: []*FieldDef{
				{
					Name: "audio",
					Type: audio,
				},
			},
			Extensions: &TypeDefExtensions{
				PublicExports: []extensions.PublicExport{
					{Group: "chat.completions", Name: "chat_completion_audio"},
				},
			},
		}

		a := publicExportTestAST(wrapper, audio)
		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.Groups, 2)
		completionsGroup := result.Groups[1]
		require.Len(t, completionsGroup.Exports, 1)
		assert.Equal(t, "chat_completion_audio", completionsGroup.Exports[0].Name)
		assert.Same(t, audio, completionsGroup.Exports[0].Target)
	})

	t.Run("drops conflicting export names in same group", func(t *testing.T) {
		t.Parallel()

		first := publicExportTestType("First")
		first.Extensions = &TypeDefExtensions{
			PublicExports: []extensions.PublicExport{
				{Group: "chat.completions", Name: "chat_completion"},
			},
		}

		second := publicExportTestType("Second")
		second.Extensions = &TypeDefExtensions{
			PublicExports: []extensions.PublicExport{
				{Group: "chat.completions", Name: "chat_completion"},
			},
		}

		a := publicExportTestAST(nil, first, second)
		a.Components.Set(first.GetRegistrationID(), first)
		a.Components.Set(second.GetRegistrationID(), second)

		result := BuildPublicExports(t.Context(), a)

		assert.Nil(t, result)
	})

	t.Run("keeps least specific duplicate target", func(t *testing.T) {
		t.Parallel()

		source := publicExportTestType("Audio")
		source.Extensions = &TypeDefExtensions{
			PublicExports: []extensions.PublicExport{
				{Group: "chat.completions", Name: "chat_completion_audio"},
			},
		}

		duplicate := publicExportTestType("Audio")
		duplicate.ContextStack = append(duplicate.ContextStack, ContextFrame{
			Type:       ContextTypeConstProperty,
			Identifier: "Assistant",
		})
		duplicate.Extensions = source.Extensions.Clone()

		a := publicExportTestAST(nil, source, duplicate)
		a.Components.Set(source.GetRegistrationID(), source)
		a.Components.Set(duplicate.GetRegistrationID(), duplicate)

		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.Groups, 2)
		completionsGroup := result.Groups[1]
		require.Len(t, completionsGroup.Exports, 1)
		assert.Same(t, source, completionsGroup.Exports[0].Target)
	})

	t.Run("input export prefers request flavor", func(t *testing.T) {
		t.Parallel()

		output := publicExportTestType("AudioContent")
		output.Extensions = &TypeDefExtensions{
			PublicExports: []extensions.PublicExport{
				{Group: "records", Name: "audio_content_param", Representation: extensions.PublicExportRepresentationInput},
			},
		}

		input := publicExportTestType("ContentAudioContent")
		input.UsedInRequest = true
		input.ContextStack = append(input.ContextStack, ContextFrame{
			Type:       ContextTypeProperty,
			Identifier: "content",
		})
		input.Extensions = output.Extensions.Clone()

		a := publicExportTestAST(nil, output, input)
		a.Components.Set(output.GetRegistrationID(), output)
		a.Components.Set(input.GetRegistrationID(), input)

		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.Groups, 1)
		recordsGroup := result.Groups[0]
		require.Len(t, recordsGroup.Exports, 1)
		assert.Same(t, input, recordsGroup.Exports[0].Target)
		assert.True(t, recordsGroup.Exports[0].Input)
	})

	t.Run("prefers target matching requested export name", func(t *testing.T) {
		t.Parallel()

		processor := publicExportTestType("Processor")
		processor.Type = DataTypeUnion
		processor.Extensions = &TypeDefExtensions{
			PublicExports: []extensions.PublicExport{
				{Group: "records", Name: "processor"},
			},
		}

		handler := publicExportTestType("Handler")
		handler.Extensions = processor.Extensions.Clone()
		processor.AssociatedTypes = TypeDefs{handler}

		a := publicExportTestAST(processor, processor, handler)
		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.Groups, 1)
		recordsGroup := result.Groups[0]
		require.Len(t, recordsGroup.Exports, 1)
		assert.Equal(t, "processor", recordsGroup.Exports[0].Name)
		assert.Same(t, processor, recordsGroup.Exports[0].Target)
	})

	t.Run("preserves input flag when preferred target changes", func(t *testing.T) {
		t.Parallel()

		handler := publicExportTestType("Handler")
		handler.Extensions = &TypeDefExtensions{
			PublicExports: []extensions.PublicExport{
				{Group: "records", Name: "processor_param"},
			},
		}

		processor := publicExportTestType("ProcessorParam")
		processor.Extensions = &TypeDefExtensions{
			PublicExports: []extensions.PublicExport{
				{Group: "records", Name: "processor_param", Representation: extensions.PublicExportRepresentationInput},
			},
		}

		a := publicExportTestAST(nil, handler, processor)
		a.Components.Set(handler.GetRegistrationID(), handler)
		a.Components.Set(processor.GetRegistrationID(), processor)

		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.Groups, 1)
		recordsGroup := result.Groups[0]
		require.Len(t, recordsGroup.Exports, 1)
		assert.Same(t, processor, recordsGroup.Exports[0].Target)
		assert.True(t, recordsGroup.Exports[0].Input)
	})

	t.Run("duplicate typed dict export upgrades same target", func(t *testing.T) {
		t.Parallel()

		processor := publicExportTestType("Processor")
		processor.Extensions = &TypeDefExtensions{
			PublicExports: []extensions.PublicExport{
				{Group: "records", Name: "processor_param"},
				{Group: "records", Name: "processor_param", Representation: extensions.PublicExportRepresentationInput},
			},
		}

		a := publicExportTestAST(processor, processor)
		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.Groups, 1)
		recordsGroup := result.Groups[0]
		require.Len(t, recordsGroup.Exports, 1)
		assert.Same(t, processor, recordsGroup.Exports[0].Target)
		assert.True(t, recordsGroup.Exports[0].Input)
	})

	t.Run("builds export from operation request type", func(t *testing.T) {
		t.Parallel()

		request := publicExportTestType("ListItemsRequest")
		request.UsedInRequest = true
		operation := Operation{
			BaseOperation: BaseOperation{
				Request: &Request{
					Field: &FieldDef{Type: request},
				},
			},
			Extensions: &OperationExtensions{
				PublicExports: []extensions.PublicExport{
					{Group: "records", Name: "item_list_params", Representation: extensions.PublicExportRepresentationInput},
				},
			},
		}

		a := publicExportTestAST(nil, request)
		a.MainSDK.AddOperations(operation)

		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.Groups, 1)
		recordsGroup := result.Groups[0]
		require.Len(t, recordsGroup.Exports, 1)
		assert.Equal(t, "item_list_params", recordsGroup.Exports[0].Name)
		assert.Same(t, request, recordsGroup.Exports[0].Target)
		assert.True(t, recordsGroup.Exports[0].Input)
	})

	t.Run("infers ambient children and suppresses implicit model namespace exports", func(t *testing.T) {
		t.Parallel()

		modelNamespace := "records"
		labelChange := publicExportTestType("LabelChange")
		labelChange.Fields = Fields{
			publicExportAmbientOptionalTestField("text", &TypeDef{Type: DataTypeString}),
		}
		entryChangeData := publicExportTestType("EntryChangeData")
		entryChangeData.Type = DataTypeUnion
		entryChangeData.AssociatedTypes = TypeDefs{labelChange}

		entryChange := publicExportTestType("EntryChange")
		entryChange.Fields = Fields{
			publicExportAmbientTestField("delta", entryChangeData),
		}
		entryChange.Extensions = &TypeDefExtensions{
			ModelNamespace: &modelNamespace,
			PublicExports: []extensions.PublicExport{
				{
					Group: "records",
					Name:  "EntryChange",
				},
			},
		}

		internal := publicExportTestType("InternalOnly")
		internal.Extensions = &TypeDefExtensions{
			ModelNamespace: &modelNamespace,
		}

		a := publicExportTestAST(entryChange, entryChange, entryChangeData, labelChange, internal)
		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.Groups, 2)

		recordsGroup := result.Groups[0]
		assert.Equal(t, "records", recordsGroup.Group)
		require.Len(t, recordsGroup.Exports, 1)
		assert.Equal(t, "EntryChange", recordsGroup.Exports[0].Name)
		assert.Same(t, entryChange, recordsGroup.Exports[0].Target)
		require.Len(t, recordsGroup.Children, 1)
		assert.Equal(t, "records.EntryChange", recordsGroup.Children[0].Group)

		entryChangeGroup := result.Groups[1]
		assert.Equal(t, "records.EntryChange", entryChangeGroup.Group)
		require.Len(t, entryChangeGroup.Exports, 1)
		assert.Equal(t, "Label", entryChangeGroup.Exports[0].Name)
		assert.Same(t, labelChange, entryChangeGroup.Exports[0].Target)
	})

	t.Run("keeps implicit model namespace exports without ambient children", func(t *testing.T) {
		t.Parallel()

		modelNamespace := "webhooks"
		webhook := publicExportTestType("Webhook")
		webhook.Extensions = &TypeDefExtensions{
			ModelNamespace: &modelNamespace,
			PublicExports: []extensions.PublicExport{
				{
					Group: "webhooks",
					Name:  "Webhook",
				},
			},
		}

		pingRequest := publicExportTestType("PingWebhookRequest")
		pingRequest.Extensions = &TypeDefExtensions{
			ModelNamespace: &modelNamespace,
		}

		a := publicExportTestAST(webhook, webhook, pingRequest)
		result := BuildPublicExports(t.Context(), a)

		require.NotNil(t, result)
		require.Len(t, result.Groups, 1)

		webhooksGroup := result.Groups[0]
		assert.Equal(t, "webhooks", webhooksGroup.Group)
		require.Len(t, webhooksGroup.Exports, 2)
		assert.Equal(t, "PingWebhookRequest", webhooksGroup.Exports[0].Name)
		assert.True(t, webhooksGroup.Exports[0].Implicit)
		assert.Equal(t, "Webhook", webhooksGroup.Exports[1].Name)
		assert.False(t, webhooksGroup.Exports[1].Implicit)
	})
}

type publicExportShape map[string]publicExportShape

func TestPublicExportInferAmbientExportsFromResourceModelGraph(t *testing.T) {
	t.Parallel()

	entry := publicExportAmbientTestType("Entry", DataTypeUnion)
	draftEntry := publicExportAmbientTestType("DraftEntry", DataTypeClass)
	finalEntry := publicExportAmbientTestType("FinalEntry", DataTypeClass)
	entry.AssociatedTypes = TypeDefs{draftEntry, finalEntry}

	labelChange := publicExportAmbientTestType("LabelChange", DataTypeClass)
	colorChange := publicExportAmbientTestType("ColorChange", DataTypeClass)
	entryChangeData := publicExportAmbientTestType("EntryChangeData", DataTypeUnion)
	entryChangeData.AssociatedTypes = TypeDefs{labelChange, colorChange}
	pageMetadata := publicExportAmbientTestType("PageMetadata", DataTypeClass)
	entryChange := publicExportAmbientTestType("EntryChange", DataTypeClass)
	entryChange.Fields = Fields{
		publicExportAmbientTestField("delta", entryChangeData),
		publicExportAmbientTestField("metadata", pageMetadata),
	}

	record := publicExportAmbientTestType("Record", DataTypeClass)
	recordSSEEventRecord := publicExportAmbientTestType("RecordSSEEventRecord", DataTypeClass)
	recordCreatedEvent := publicExportAmbientTestType("RecordCreatedEvent", DataTypeClass)
	recordCreatedEvent.Fields = Fields{
		publicExportAmbientTestField("record", recordSSEEventRecord),
		publicExportAmbientTestField("metadata", pageMetadata),
	}

	recordGetParamsNonStreaming := publicExportAmbientTestType("RecordGetParamsNonStreaming", DataTypeClass)
	recordGetParamsStreaming := publicExportAmbientTestType("RecordGetParamsStreaming", DataTypeClass)
	recordGetParams := publicExportAmbientTestType("RecordGetParams", DataTypeUnion)
	recordGetParams.AssociatedTypes = TypeDefs{recordGetParamsNonStreaming, recordGetParamsStreaming}

	labelPayloadParam := publicExportAmbientTestType("CreateRecordLabelPayloadParam", DataTypeClass)
	labelPayloadParam.Fields = Fields{
		publicExportAmbientTypeConstTestField("label"),
		publicExportAmbientOptionalTestField("label", &TypeDef{Type: DataTypeString}),
	}
	colorPayloadParam := publicExportAmbientTestType("CreateRecordColorPayloadParam", DataTypeClass)
	colorPayloadParam.Fields = Fields{
		publicExportAmbientTypeConstTestField("color"),
		publicExportAmbientOptionalTestField("url", &TypeDef{Type: DataTypeString}),
	}
	createRecordPayloadParam := publicExportAmbientTestType("CreateRecordPayloadParam", DataTypeUnion)
	createRecordPayloadParam.AssociatedTypes = TypeDefs{labelPayloadParam, colorPayloadParam}
	createAgentRecord := publicExportAmbientTestType("CreateDraftRecord", DataTypeClass)
	createAgentRecord.Fields = Fields{
		publicExportAmbientTestField("payload", createRecordPayloadParam),
	}
	createModelRecord := publicExportAmbientTestType("CreateFinalRecord", DataTypeClass)
	recordCreateParams := publicExportAmbientTestType("RecordCreateParams", DataTypeUnion)
	recordCreateParams.AssociatedTypes = TypeDefs{createAgentRecord, createModelRecord}

	handler := publicExportAmbientTestType("Handler", DataTypeClass)
	calculation := publicExportAmbientTestType("Calculation", DataTypeClass)
	indexedLookupOptions := publicExportAmbientTestType("IndexedLookupOptions", DataTypeClass)
	indexedLookupOptions.Fields = Fields{
		publicExportAmbientTestField("engine", &TypeDef{Type: DataTypeString}),
	}
	remoteLookupOptions := publicExportAmbientTestType("RemoteLookupOptions", DataTypeClass)
	remoteLookupOptions.Fields = Fields{
		publicExportAmbientTestField("api_key", &TypeDef{Type: DataTypeString}),
		publicExportAmbientOptionalTestField("provider_settings", &TypeDef{Type: DataTypeMap, ItemType: &TypeDef{Type: DataTypeAny}}),
	}
	looseLookupOptions := publicExportAmbientTestType("LooseLookupOptions", DataTypeClass)
	looseLookupOptions.Fields = Fields{
		publicExportAmbientOptionalTestField("provider_settings", &TypeDef{Type: DataTypeMap, ItemType: &TypeDef{Type: DataTypeAny}}),
	}
	storedRecord := publicExportAmbientTestType("StoredRecord", DataTypeClass)
	storageOptions := publicExportAmbientTestType("StorageOptions", DataTypeClass)
	storageOptions.Fields = Fields{
		publicExportAmbientOptionalTestField("stored_records", publicExportAmbientTestArray(storedRecord)),
	}
	lookup := publicExportAmbientTestType("Lookup", DataTypeClass)
	lookup.Fields = Fields{
		publicExportAmbientTestField("remote_lookup_options", remoteLookupOptions),
		publicExportAmbientTestField("loose_lookup_options", looseLookupOptions),
		publicExportAmbientTestField("storage_options", storageOptions),
		publicExportAmbientTestField("indexed_lookup_options", indexedLookupOptions),
	}
	processor := publicExportAmbientTestType("Processor", DataTypeUnion)
	processor.AssociatedTypes = TypeDefs{handler, calculation, lookup}

	calculationCallArguments := publicExportAmbientTestType("CalculationCallArguments", DataTypeClass)
	calculationCallEntry := publicExportAmbientTestType("CalculationCallEntry", DataTypeClass)
	calculationCallEntry.Fields = Fields{
		publicExportAmbientTypeConstTestField("calculation_call"),
		publicExportAmbientTestField("arguments", calculationCallArguments),
	}

	stockNote := publicExportAmbientTestType("StockNote", DataTypeClass)
	item := publicExportAmbientTestType("CatalogResultItems", DataTypeClass)
	item.Fields = Fields{
		publicExportAmbientTestField("stock_notes", publicExportAmbientTestArray(stockNote)),
	}
	catalogResult := publicExportAmbientTestType("CatalogResult", DataTypeClass)
	catalogResult.Fields = Fields{
		publicExportAmbientTestField("items", publicExportAmbientTestArray(item)),
	}
	catalogResultEntry := publicExportAmbientTestType("CatalogResultEntry", DataTypeClass)
	catalogResultEntry.Fields = Fields{
		publicExportAmbientTypeConstTestField("catalog_result"),
		publicExportAmbientTestField("result", publicExportAmbientTestArray(catalogResult)),
	}

	exports := publicExportInferAmbientExports([]publicExportAmbientRoot{
		{Group: "records", Name: "CalculationCallArguments", Target: calculationCallArguments},
		{Group: "records", Name: "CalculationCallEntry", Target: calculationCallEntry},
		{Group: "records", Name: "CreateDraftRecordParamsStreaming", Target: createAgentRecord, Input: true},
		{Group: "records", Name: "CreateFinalRecordParamsStreaming", Target: createModelRecord, Input: true},
		{Group: "records", Name: "Handler", Target: handler},
		{Group: "records", Name: "CatalogResult", Target: catalogResult},
		{Group: "records", Name: "CatalogResultEntry", Target: catalogResultEntry},
		{Group: "records", Name: "Record", Target: record},
		{Group: "records", Name: "RecordCreateParams", Target: recordCreateParams, Input: true},
		{Group: "records", Name: "RecordCreatedEvent", Target: recordCreatedEvent},
		{Group: "records", Name: "RecordGetParams", Target: recordGetParams, Input: true},
		{Group: "records", Name: "RecordGetParamsNonStreaming", Target: recordGetParamsNonStreaming, Input: true},
		{Group: "records", Name: "RecordGetParamsStreaming", Target: recordGetParamsStreaming, Input: true},
		{Group: "records", Name: "FinalEntry", Target: finalEntry},
		{Group: "records", Name: "Entry", Target: entry},
		{Group: "records", Name: "EntryChange", Target: entryChange},
		{Group: "records", Name: "Processor", Target: processor},
		{Group: "records", Name: "DraftEntry", Target: draftEntry},
	})

	actual := publicExportAmbientDebugLines(exports)
	t.Logf("ambient nested exports:\n%s", strings.Join(actual, "\n"))

	expected := []string{
		"records.CalculationCallEntry.Arguments -> CalculationCallArguments [field arguments]",
		"records.CatalogResult.Item -> CatalogResultItems [field items]",
		"records.CatalogResult.Item.StockNote -> StockNote [field stock_notes]",
		"records.CatalogResultEntry.Result -> CatalogResult [field result]",
		"records.CatalogResultEntry.Result.Item -> CatalogResultItems [field items]",
		"records.CatalogResultEntry.Result.Item.StockNote -> StockNote [field stock_notes]",
		"records.EntryChange.Color -> ColorChange [oneOf EntryChangeData]",
		"records.EntryChange.Label -> LabelChange [oneOf EntryChangeData]",
		"records.EntryChange.Metadata -> PageMetadata [field metadata]",
		"records.Processor.Calculation -> Calculation [oneOf Processor]",
		"records.Processor.Lookup -> Lookup [oneOf Processor]",
		"records.Processor.Lookup.IndexedLookupOptions -> IndexedLookupOptions [field indexed_lookup_options]",
		"records.RecordCreatedEvent.Metadata -> PageMetadata [field metadata]",
		"records.RecordGetParams.RecordGetParamsNonStreaming -> RecordGetParamsNonStreaming [oneOf RecordGetParams]",
		"records.RecordGetParams.RecordGetParamsStreaming -> RecordGetParamsStreaming [oneOf RecordGetParams]",
	}
	assert.Equal(t, expected, actual)
	assert.NotContains(t, actual, "records.CreateDraftRecordParamsStreaming.Color -> CreateRecordColorPayloadParam [oneOf CreateRecordPayloadParam]")
	assert.NotContains(t, actual, "records.CreateDraftRecordParamsStreaming.Label -> CreateRecordLabelPayloadParam [oneOf CreateRecordPayloadParam]")
	assert.NotContains(t, actual, "records.RecordCreateParams.CreateDraftRecord -> CreateDraftRecord [oneOf RecordCreateParams]")
	assert.NotContains(t, actual, "records.RecordCreateParams.CreateFinalRecord -> CreateFinalRecord [oneOf RecordCreateParams]")
	assert.NotContains(t, actual, "records.Entry.DraftEntry -> DraftEntry [oneOf Entry]")
	assert.NotContains(t, actual, "records.RecordCreatedEvent.Record -> RecordSSEEventRecord [field record]")
	assert.NotContains(t, actual, "records.Processor.Handler -> Handler [oneOf Processor]")
	assert.NotContains(t, actual, "records.Processor.Lookup.RemoteLookupOptions -> RemoteLookupOptions [field remote_lookup_options]")
	assert.NotContains(t, actual, "records.Processor.Lookup.LooseLookupOptions -> LooseLookupOptions [field loose_lookup_options]")
	assert.NotContains(t, actual, "records.Processor.Lookup.StorageOptions -> StorageOptions [field storage_options]")

	expectedShape := publicExportShape{
		"CalculationCallArguments": publicExportShape{},
		"CalculationCallEntry": publicExportShape{
			"Arguments": publicExportShape{},
		},
		"CreateDraftRecordParamsStreaming": publicExportShape{},
		"CreateFinalRecordParamsStreaming": publicExportShape{},
		"Handler":                          publicExportShape{},
		"CatalogResult": publicExportShape{
			"Item": publicExportShape{
				"StockNote": publicExportShape{},
			},
		},
		"CatalogResultEntry": publicExportShape{
			"Result": publicExportShape{
				"Item": publicExportShape{
					"StockNote": publicExportShape{},
				},
			},
		},
		"Record":             publicExportShape{},
		"RecordCreateParams": publicExportShape{},
		"RecordCreatedEvent": publicExportShape{
			"Metadata": publicExportShape{},
		},
		"RecordGetParams": publicExportShape{
			"RecordGetParamsNonStreaming": publicExportShape{},
			"RecordGetParamsStreaming":    publicExportShape{},
		},
		"RecordGetParamsNonStreaming": publicExportShape{},
		"RecordGetParamsStreaming":    publicExportShape{},
		"FinalEntry":                  publicExportShape{},
		"Entry":                       publicExportShape{},
		"EntryChange": publicExportShape{
			"Color":    publicExportShape{},
			"Metadata": publicExportShape{},
			"Label":    publicExportShape{},
		},
		"Processor": publicExportShape{
			"Calculation": publicExportShape{},
			"Lookup": publicExportShape{
				"IndexedLookupOptions": publicExportShape{},
			},
		},
		"DraftEntry": publicExportShape{},
	}
	assert.Equal(t, expectedShape, publicExportAmbientShapeFromExports("records", []publicExportAmbientRoot{
		{Group: "records", Name: "CalculationCallArguments", Target: calculationCallArguments},
		{Group: "records", Name: "CalculationCallEntry", Target: calculationCallEntry},
		{Group: "records", Name: "CreateDraftRecordParamsStreaming", Target: createAgentRecord, Input: true},
		{Group: "records", Name: "CreateFinalRecordParamsStreaming", Target: createModelRecord, Input: true},
		{Group: "records", Name: "Handler", Target: handler},
		{Group: "records", Name: "CatalogResult", Target: catalogResult},
		{Group: "records", Name: "CatalogResultEntry", Target: catalogResultEntry},
		{Group: "records", Name: "Record", Target: record},
		{Group: "records", Name: "RecordCreateParams", Target: recordCreateParams, Input: true},
		{Group: "records", Name: "RecordCreatedEvent", Target: recordCreatedEvent},
		{Group: "records", Name: "RecordGetParams", Target: recordGetParams, Input: true},
		{Group: "records", Name: "RecordGetParamsNonStreaming", Target: recordGetParamsNonStreaming, Input: true},
		{Group: "records", Name: "RecordGetParamsStreaming", Target: recordGetParamsStreaming, Input: true},
		{Group: "records", Name: "FinalEntry", Target: finalEntry},
		{Group: "records", Name: "Entry", Target: entry},
		{Group: "records", Name: "EntryChange", Target: entryChange},
		{Group: "records", Name: "Processor", Target: processor},
		{Group: "records", Name: "DraftEntry", Target: draftEntry},
	}, exports))
}

func TestPublicExportInferAmbientExportsFromNestedExportGroup(t *testing.T) {
	t.Parallel()

	detail := publicExportAmbientTestType("Detail", DataTypeClass)
	payload := publicExportAmbientTestType("Payload", DataTypeClass)
	payload.Fields = Fields{
		publicExportAmbientTestField("detail", detail),
	}
	response := publicExportAmbientTestType("Response", DataTypeClass)
	response.Fields = Fields{
		publicExportAmbientTestField("payload", payload),
	}

	exports := publicExportInferAmbientExports([]publicExportAmbientRoot{
		{Group: "chat.completions", Name: "Response", Target: response},
	})

	actual := publicExportAmbientDebugLines(exports)
	assert.Equal(t, []string{
		"chat.completions.Response.Payload -> Payload [field payload]",
		"chat.completions.Response.Payload.Detail -> Detail [field detail]",
	}, actual)
}

func publicExportTestType(name string) *TypeDef {
	return &TypeDef{
		Name:         name,
		OriginalName: name,
		Type:         DataTypeClass,
		Scope:        ScopeShared,
		ContextStack: ContextStack{
			{Type: ContextTypeComponent, Identifier: name},
		},
	}
}

func publicExportTestAST(rootType *TypeDef, renderedTypes ...*TypeDef) *AST {
	a := NewAST()
	a.MainSDK = NewMainSDK(rootType)
	a.BucketedTypes = NewBucketedTypes()

	models := sequencedmap.New[string, TypeDefs]()
	models.Set("models", renderedTypes)
	a.BucketedTypes.Set("models", models)

	if rootType != nil {
		a.Components.Set(rootType.GetRegistrationIDOrType(), rootType)
	}

	return a
}

func publicExportAmbientTestType(name string, dataType DataType) *TypeDef {
	t := publicExportTestType(name)
	t.Type = dataType
	if dataType == DataTypeClass {
		t.Fields = Fields{
			publicExportAmbientOptionalTestField("value", &TypeDef{Type: DataTypeString}),
		}
	}
	return t
}

func publicExportAmbientTestField(name string, typ *TypeDef) *FieldDef {
	return &FieldDef{
		Name:         name,
		OriginalName: name,
		Type:         typ,
	}
}

func publicExportAmbientTypeConstTestField(value string) *FieldDef {
	field := publicExportAmbientTestField("type", &TypeDef{Type: DataTypeString})
	field.Const = &AnyValue{Value: value}
	return field
}

func publicExportAmbientOptionalTestField(name string, typ *TypeDef) *FieldDef {
	field := publicExportAmbientTestField(name, typ)
	field.Optional = true
	return field
}

func publicExportAmbientTestArray(itemType *TypeDef) *TypeDef {
	return &TypeDef{
		Type:     DataTypeArray,
		ItemType: itemType,
	}
}

func publicExportAmbientDebugLines(exports []publicExportAmbientExport) []string {
	result := make([]string, 0, len(exports))
	for _, export := range exports {
		targetName := ""
		if export.Target != nil {
			targetName = export.Target.Name
		}
		result = append(result, export.Group+"."+export.Name+" -> "+targetName+" ["+export.Via+"]")
	}
	return result
}

func publicExportAmbientShapeFromExports(namespace string, roots []publicExportAmbientRoot, exports []publicExportAmbientExport) publicExportShape {
	shape := publicExportShape{}
	for _, root := range roots {
		if root.Group != namespace || root.Name == "" {
			continue
		}
		shape.add([]string{root.Name})
	}

	namespaceParts := publicExportGroupParts(namespace)
	for _, export := range exports {
		groupParts := publicExportGroupParts(export.Group)
		if len(groupParts) < len(namespaceParts) {
			continue
		}
		if strings.Join(groupParts[:len(namespaceParts)], ".") != strings.Join(namespaceParts, ".") {
			continue
		}
		shape.add(append(slices.Clone(groupParts[len(namespaceParts):]), export.Name))
	}
	return shape
}

func (s publicExportShape) add(path []string) {
	if len(path) == 0 {
		return
	}
	child := s[path[0]]
	if child == nil {
		child = publicExportShape{}
		s[path[0]] = child
	}
	child.add(path[1:])
}

func publicExportShapeNestedPaths(shape publicExportShape) []string {
	return publicExportShapePaths(shape, nil, true)
}

func publicExportShapePaths(shape publicExportShape, prefix []string, nestedOnly bool) []string {
	var paths []string
	keys := make([]string, 0, len(shape))
	for key := range shape {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	for _, key := range keys {
		path := append(slices.Clone(prefix), key)
		if !nestedOnly || len(path) > 1 {
			paths = append(paths, strings.Join(path, "."))
		}
		paths = append(paths, publicExportShapePaths(shape[key], path, nestedOnly)...)
	}
	return paths
}
