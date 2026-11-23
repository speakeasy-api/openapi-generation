#nullable enable
using Openapi.Models.Errors;
using Openapi.Models.Shared;
using Openapi.Utils.Json;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Xunit;

public class UnionConstShould
{
    [Fact]
    public void SelectVariantMatchingConstInEitherDeclarationOrder()
    {
        var a = JsonSerialization.Deserialize<KindChoice>("""{"kind":"a","name":"n"}""")!;
        var b = JsonSerialization.Deserialize<KindChoice>("""{"kind":"b","name":"n"}""")!;
        var reversedA = JsonSerialization.Deserialize<KindChoiceReversed>("""{"kind":"a","name":"n"}""")!;
        var reversedB = JsonSerialization.Deserialize<KindChoiceReversed>("""{"kind":"b","name":"n"}""")!;

        Assert.NotNull(a.KindA);
        Assert.NotNull(b.KindB);
        Assert.NotNull(reversedA.KindA);
        Assert.NotNull(reversedB.KindB);
        AssertJson("""{"kind":"a","mode":"fast","name":"n"}""", JsonSerialization.Serialize(a));
        AssertJson("""{"kind":"b","name":"n"}""", JsonSerialization.Serialize(b));
        AssertJson("""{"kind":"a","mode":"fast","name":"n"}""", JsonSerialization.Serialize(reversedA));
        AssertJson("""{"kind":"b","name":"n"}""", JsonSerialization.Serialize(reversedB));
    }

    [Fact]
    public void KeepUnmatchedConstAsUnknownInOpenUnion()
    {
        var json = """{"kind":"z","name":"n"}""";
        var union = JsonSerialization.Deserialize<KindChoice>(json)!;

        Assert.True(union.IsUnknown());
        AssertJson(json, JsonSerialization.Serialize(union));
    }

    [Fact]
    public void RejectUnmatchedConstInClosedUnion()
    {
        Assert.NotNull(JsonSerialization.Deserialize<ClosedKindChoice>("""{"kind":"b"}""")!.ClosedKindB);
        Assert.Throws<JsonException>(() => JsonSerialization.Deserialize<ClosedKindChoice>("""{"kind":"z"}"""));
    }

    [Theory]
    [InlineData("""{"kind":"z","name":"n"}""")]
    [InlineData("""{"kind":null,"name":"n"}""")]
    public void FallBackToMapWhenNoConstMatches(string json)
    {
        var union = JsonSerialization.Deserialize<KindOrMap>(json)!;

        Assert.NotNull(union.MapOfAny);
        AssertJson(json, JsonSerialization.Serialize(union));
    }

    [Fact]
    public void KeepDeclaredSpecificityWithUndeclaredKey()
    {
        var union = JsonSerialization.Deserialize<KindOrMap>("""{"kind":"b","name":"n","extra":true}""")!;

        Assert.NotNull(union.KindB);
    }

    [Theory]
    [InlineData("""{"kind":"a","kind":"b","name":"n"}""")]
    [InlineData("""{"kind":"b","kind":"a","name":"n"}""")]
    public void RejectCandidateWithAnyMismatchedRepeatedConst(string json)
    {
        Assert.NotNull(JsonSerialization.Deserialize<KindOrMap>(json)!.MapOfAny);
    }

    [Fact]
    public void AcceptRepeatedMatchingConst()
    {
        Assert.NotNull(JsonSerialization.Deserialize<KindOrMap>("""{"kind":"b","kind":"b","name":"n"}""")!.KindB);
    }

    [Fact]
    public void KeepDefaultsForMissingConsts()
    {
        var union = JsonSerialization.Deserialize<KindOrMap>("""{"name":"n"}""")!;

        Assert.NotNull(union.KindB);
        Assert.Equal("b", union.KindB.Kind);
        Assert.Equal("c", JsonSerialization.Deserialize<OptionalKind>("{}")!.Kind);
        Assert.Equal("d", JsonSerialization.Deserialize<DefaultedKind>("{}")!.Kind);
        Assert.Equal("e", JsonSerialization.Deserialize<ConflictingDefaultKind>("{}")!.Kind);
    }

    [Fact]
    public void KeepDirectModelReadsUnchanged()
    {
        Assert.Equal("a", JsonSerialization.Deserialize<KindA>("""{"kind":"b"}""")!.Kind);
        Assert.Equal(8, JsonSerialization.Deserialize<CodedFailureAPayload>("""{"code":8,"detail":{"level":1}}""")!.Code);
    }

    [Fact]
    public void PreferConstOverDefault()
    {
        Assert.NotNull(JsonSerialization.Deserialize<ConflictingDefaultKindOrMap>("""{"kind":"e"}""")!.ConflictingDefaultKind);
        Assert.NotNull(JsonSerialization.Deserialize<ConflictingDefaultKindOrMap>("""{"kind":"z"}""")!.MapOfAny);
    }

    [Fact]
    public void KeepIncomingValueOfDefaultedProperty()
    {
        var union = JsonSerialization.Deserialize<KindChoice>("""{"kind":"a","mode":"slow"}""")!;

        Assert.NotNull(union.KindA);
        Assert.Equal("slow", union.KindA.Mode);
    }

    [Theory]
    [InlineData("""{"flag":true}""", true)]
    [InlineData("""{"flag":false}""", false)]
    public void SelectBooleanConst(string json, bool flag)
    {
        var union = JsonSerialization.Deserialize<FlagChoice>(json)!;

        Assert.Equal(flag, union.FlagOn != null);
        Assert.Equal(!flag, union.FlagOff != null);
    }

    [Theory]
    [InlineData("""{"flag":1}""")]
    [InlineData("""{"flag":"true"}""")]
    [InlineData("""{"flag":null}""")]
    public void RejectBooleanConstOfOtherKind(string json)
    {
        Assert.NotNull(JsonSerialization.Deserialize<FlagChoice>(json)!.MapOfAny);
    }

    [Theory]
    [InlineData("""{"level":1}""")]
    [InlineData("""{"level":1.0}""")]
    [InlineData("""{"level":1e0}""")]
    [InlineData("""{"level":10E-1}""")]
    [InlineData("""{"level":0.001e3}""")]
    public void MatchIntegerConstInEquivalentNumberForms(string json)
    {
        var union = JsonSerialization.Deserialize<LevelChoice>(json)!;

        Assert.NotNull(union.LevelOne);
        Assert.Null(union.LevelZero);
    }

    [Theory]
    [InlineData("""{"level":0}""")]
    [InlineData("""{"level":-0}""")]
    [InlineData("""{"level":0.0}""")]
    [InlineData("""{"level":-0e-5}""")]
    [InlineData("""{"level":0e99999999999999999999}""")]
    public void MatchZeroConstInEveryZeroForm(string json)
    {
        var union = JsonSerialization.Deserialize<LevelChoice>(json)!;

        Assert.NotNull(union.LevelZero);
        Assert.Null(union.LevelOne);
    }

    [Theory]
    [InlineData("""{"level":"1"}""")]
    [InlineData("""{"level":true}""")]
    [InlineData("""{"level":-1}""")]
    [InlineData("""{"level":1.0000000000000000001}""")]
    [InlineData("""{"level":1e99999999999999999999}""")]
    public void RejectIntegerConstOfOtherValueOrKind(string json)
    {
        var union = JsonSerialization.Deserialize<LevelChoice>(json)!;

        Assert.Null(union.LevelZero);
        Assert.Null(union.LevelOne);
    }

    [Theory]
    [InlineData("""{"ratio":2.5}""", true)]
    [InlineData("""{"ratio":25e-1}""", true)]
    [InlineData("""{"ratio":2.50}""", true)]
    [InlineData("""{"ratio":1.5}""", false)]
    public void SelectNumberConst(string json, bool high)
    {
        var union = JsonSerialization.Deserialize<RatioChoice>(json)!;

        Assert.Equal(high, union.RatioHigh != null);
        Assert.Equal(!high, union.RatioLow != null);
    }

    [Theory]
    [InlineData("""{"ratio":"2.5"}""")]
    [InlineData("""{"ratio":2.5000000000000000001}""")]
    public void RejectNumberConstOfOtherValueOrKind(string json)
    {
        var union = JsonSerialization.Deserialize<RatioChoice>(json)!;

        Assert.Null(union.RatioLow);
        Assert.Null(union.RatioHigh);
    }

    [Theory]
    [InlineData("""{"text":"say \"hi\" \\ é"}""", true)]
    [InlineData("""{"text":"say \u0022hi\u0022 \\ \u00e9"}""", true)]
    [InlineData("""{"text":"say \"hi\" \\ e"}""", false)]
    [InlineData("""{"text":"say \"hi\" \\\\ é"}""", false)]
    public void CompareDecodedStringConst(string json, bool matches)
    {
        var union = JsonSerialization.Deserialize<QuotedTextOrMap>(json)!;

        Assert.Equal("say \"hi\" \\ é", new QuotedText().Text);
        Assert.Equal(matches, union.QuotedText != null);
        Assert.Equal(!matches, union.MapOfAny != null);
    }

    [Theory]
    [InlineData("""{"at":"2020-01-01T00:00:00Z"}""", true)]
    [InlineData("""{"at":"2020-01-01T00:00:00Z","day":"2020-01-02"}""", true)]
    [InlineData("""{"at":"2020-01-01T00:00:00.000Z"}""", false)]
    [InlineData("""{"at":"2020-01-01T00:00:00+00:00"}""", false)]
    [InlineData("""{"at":"2020-01-01T00:00:00Z","day":"2020-01-03"}""", false)]
    public void CompareTemporalConstAsLiteralString(string json, bool stamped)
    {
        var union = JsonSerialization.Deserialize<StampedOrMap>(json)!;

        Assert.Equal(stamped, union.Stamped != null);
        Assert.Equal(!stamped, union.MapOfAny != null);
    }

    [Theory]
    [InlineData("""{"child":{"kind":"a"}}""")]
    [InlineData("""{"items":[{"kind":"a"}]}""")]
    [InlineData("""{"entries":{"x":{"kind":"a"}}}""")]
    [InlineData("""{"maybe":{"kind":"a"}}""")]
    public void SelectCandidateWhoseNestedConstMatches(string json)
    {
        Assert.NotNull(JsonSerialization.Deserialize<KindHolderChoice>(json)!.KindHolderA);
    }

    [Theory]
    [InlineData("""{"child":{"kind":"b"}}""")]
    [InlineData("""{"items":[{"kind":"b"}]}""")]
    [InlineData("""{"entries":{"x":{"kind":"b"}}}""")]
    [InlineData("""{"maybe":{"kind":"b"}}""")]
    public void RejectCandidateWithMismatchedNestedConst(string json)
    {
        var union = JsonSerialization.Deserialize<KindHolderChoice>(json)!;

        Assert.NotNull(union.KindHolderB);
        AssertJson(json, JsonSerialization.Serialize(union));
    }

    [Theory]
    [InlineData("""{"items":[{"kind":"a"},{"kind":"b"}]}""")]
    [InlineData("""{"entries":{"x":{"kind":"a"},"y":{"kind":"b"}}}""")]
    public void RejectEveryCandidateWithMixedNestedConsts(string json)
    {
        var union = JsonSerialization.Deserialize<KindHolderChoice>(json)!;

        Assert.NotNull(union.MapOfAny);
        AssertJson(json, JsonSerialization.Serialize(union));
    }

    [Theory]
    [InlineData("""{"child":{"kind":"a"},"child":{"kind":"b"}}""")]
    [InlineData("""{"child":{"kind":"b"},"child":{"kind":"a"}}""")]
    public void RejectCandidateWithAnyMismatchedRepeatedChild(string json)
    {
        Assert.NotNull(JsonSerialization.Deserialize<KindHolderChoice>(json)!.MapOfAny);
    }

    [Fact]
    public void AcceptNullNestedConstHolder()
    {
        var union = JsonSerialization.Deserialize<KindHolderChoice>("""{"child":{"kind":"a"},"maybe":null}""")!;

        Assert.NotNull(union.KindHolderA);
        Assert.True(union.KindHolderA.Maybe.IsNull);
    }

    [Theory]
    [InlineData("a", "a", true)]
    [InlineData("b", "b", false)]
    [InlineData("a", "b", null)]
    public void ValidateConstsBeyondScoringDepth(string outer, string innermost, bool? expectA)
    {
        var json = NestedNodes(12, outer, innermost);
        var union = JsonSerialization.Deserialize<KindNodeChoice>(json)!;

        Assert.Equal(expectA == true, union.KindNodeA != null);
        Assert.Equal(expectA == false, union.KindNodeB != null);
        Assert.Equal(expectA == null, union.MapOfAny != null);
    }

    [Fact]
    public void ScoreSelectedInnerVariantOfNestedUnion()
    {
        var json = """{"value":{"kind":"b","name":"n","note":"x"},"alive":true}""";
        var inner = JsonSerialization.Deserialize<InnerKindChoice>("""{"kind":"b","name":"n","note":"x"}""")!;
        var union = JsonSerialization.Deserialize<WrappedKindChoice>(json)!;

        Assert.NotNull(inner.LeanKindB);
        Assert.NotNull(union.LooseNoteWrapper);
        Assert.True(union.LooseNoteWrapper.Alive);
        Assert.Equal("x", union.LooseNoteWrapper.Value.Note);
    }

    [Fact]
    public void ScoreNestedUnionByTheMatchesOfItsSelectedVariant()
    {
        var union = JsonSerialization.Deserialize<WrappedKindChoice>("""{"value":{"kind":"a","name":"n","note":"x"}}""")!;

        Assert.NotNull(union.InnerKindWrapper);
        Assert.NotNull(union.InnerKindWrapper.Value.RichKindA);
    }

    [Theory]
    [InlineData("""{"kinds":[{"kind":"a","name":"n","note":"x"}]}""", true)]
    [InlineData("""{"kinds":[{"kind":"b","name":"n","note":"x"}],"alive":true}""", false)]
    [InlineData("""{"riches":[{"name":"item","meta":{"a":"q","b":"r"}}],"alive":true}""", false)]
    public void ScoreNestedUnionListItemsBySelectedVariant(string json, bool kindChoice)
    {
        var union = JsonSerialization.Deserialize<ListHolderChoice>(json)!;

        Assert.Equal(kindChoice, union.KindChoiceListHolder != null);
        Assert.Equal(!kindChoice, union.LooseListHolder != null);
        if (kindChoice)
        {
            Assert.NotNull(union.KindChoiceListHolder!.Kinds![0].RichKindA);
            AssertJson(json, JsonSerialization.Serialize(union));
        }
        else
        {
            Assert.True(union.LooseListHolder!.Alive);
        }
    }

    [Theory]
    [InlineData("""{"kinds":{"x":{"kind":"a","name":"n","note":"x"}}}""", true)]
    [InlineData("""{"kinds":{"x":{"kind":"b","name":"n","note":"x"}},"alive":true}""", false)]
    [InlineData("""{"riches":{"x":{"name":"item","meta":{"a":"q","b":"r"}}},"alive":true}""", false)]
    public void ScoreNestedUnionMapValuesBySelectedVariant(string json, bool kindChoice)
    {
        var union = JsonSerialization.Deserialize<MapHolderChoice>(json)!;

        Assert.Equal(kindChoice, union.KindChoiceMapHolder != null);
        Assert.Equal(!kindChoice, union.LooseMapHolder != null);
        if (kindChoice)
        {
            Assert.NotNull(union.KindChoiceMapHolder!.Kinds!["x"].RichKindA);
            AssertJson(json, JsonSerialization.Serialize(union));
        }
        else
        {
            Assert.True(union.LooseMapHolder!.Alive);
        }
    }

    [Fact]
    public void ScoreRejectedInnerVariantAsItsActualFallback()
    {
        var json = """{"value":{"name":"item","meta":{"a":"q","b":"r"}},"alive":true}""";
        var inner = JsonSerialization.Deserialize<RichOrMap>("""{"name":"item","meta":{"a":"q","b":"r"}}""")!;
        var union = JsonSerialization.Deserialize<WrapperChoice>(json)!;

        Assert.NotNull(inner.MapOfAny);
        Assert.NotNull(union.TypedWrapper);
        AssertJson(json, JsonSerialization.Serialize(union));
    }

    [Fact]
    public void PenalizeUnknownInnerUnion()
    {
        var json = """{"value":{"kind":"z"}}""";
        var holder = JsonSerialization.Deserialize<OpenKindHolder>(json)!;
        var union = JsonSerialization.Deserialize<OpenKindHolderOrMap>(json)!;

        Assert.True(holder.Value.IsUnknown());
        AssertJson(json, JsonSerialization.Serialize(holder));
        Assert.NotNull(union.MapOfAny);
        AssertJson(json, JsonSerialization.Serialize(union));
    }

    [Theory]
    [InlineData("""{"item":{"tag":"a","level":1}}""", true)]
    [InlineData("""{"item":{"tag":"b","level":1}}""", false)]
    [InlineData("""{"item":{"tag":"a","level":2}}""", false)]
    [InlineData("""{"item":{"tag":"z","level":5}}""", true)]
    public void ValidateConstsOfSelectedDiscriminatedVariant(string json, bool holder)
    {
        var union = JsonSerialization.Deserialize<TaggedLevelHolderOrMap>(json)!;

        Assert.Equal(holder, union.TaggedLevelHolder != null);
        Assert.Equal(!holder, union.MapOfAny != null);
        AssertJson(json, JsonSerialization.Serialize(union));
    }

    [Theory]
    [InlineData("""{"code":7,"detail":{"level":1}}""", true)]
    [InlineData("""{"code":7.0,"detail":{"level":1}}""", true)]
    [InlineData("""{"code":7e0,"detail":{"level":2}}""", false)]
    [InlineData("""{"code":0.7e1,"detail":{"level":2}}""", false)]
    public void BindSettableIntegralConstInEquivalentNumberForms(string json, bool first)
    {
        var union = JsonSerialization.Deserialize<ConstUnionChoicesGetResponseBody>(json)!;

        Assert.Equal(first, union.CodedFailureAPayload != null);
        Assert.Equal(!first, union.CodedFailureBPayload != null);
        Assert.Equal(7, first ? union.CodedFailureAPayload!.Code : union.CodedFailureBPayload!.Code);
        AssertJson(first ? """{"code":7,"detail":{"level":1}}""" : """{"code":7,"detail":{"level":2}}""", JsonSerialization.Serialize(union));
    }

    [Theory]
    [InlineData("""{"code":"7","detail":{"level":1}}""")]
    [InlineData("""{"code":8,"detail":{"level":1}}""")]
    [InlineData("""{"code":7,"detail":{"level":3}}""")]
    public void RejectSettableConstOfOtherValueOrKind(string json)
    {
        var union = JsonSerialization.Deserialize<ConstUnionChoicesGetResponseBody>(json)!;

        Assert.True(union.IsUnknown());
        AssertJson(json, JsonSerialization.Serialize(union));
    }

    private static string NestedNodes(int depth, string outer, string innermost)
    {
        var json = new StringBuilder();
        for (var i = 0; i < depth; i++)
        {
            json.Append("{\"kind\":\"").Append(i == depth - 1 ? innermost : outer).Append('"');
            if (i < depth - 1)
            {
                json.Append(",\"next\":");
            }
        }
        json.Append('}', depth);
        return json.ToString();
    }

    private static void AssertJson(string expected, string actual)
    {
        Assert.True(JsonNode.DeepEquals(JsonNode.Parse(expected), JsonNode.Parse(actual)), actual);
    }
}
