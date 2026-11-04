#nullable enable
using System;
using System.Collections.Generic;
using System.Linq;
using System.Reflection;
using Newtonsoft.Json;
using Openapi.Models.Shared;
using Openapi.Utils;
using Xunit;

public class UnionImplicitConversionsAdditionalShould
{
    private static string[] ImplicitConversionSources<TUnion>() =>
        typeof(TUnion)
            .GetMethods(BindingFlags.Public | BindingFlags.Static)
            .Where(m => m.Name == "op_Implicit" && m.ReturnType == typeof(TUnion))
            .Select(m => m.GetParameters()[0].ParameterType.Name)
            .OrderBy(name => name, StringComparer.Ordinal)
            .ToArray();

    [Fact]
    public void ConvertDiscriminatedMemberAndSetDiscriminator()
    {
        var circle = new ImplicitConversionCircle { Kind = "", Radius = 2 };

        ImplicitConversionShape shape = circle;

        Assert.Equal(ImplicitConversionShapeType.Circle, shape.Type);
        Assert.Same(circle, shape.ImplicitConversionCircle);
        Assert.Equal("circle", circle.Kind);
        Assert.Equal(
            JsonConvert.SerializeObject(
                ImplicitConversionShape.CreateCircle(
                    new ImplicitConversionCircle { Kind = "", Radius = 2 }
                ), Utilities.GetDefaultJsonSerializerSettings()),
            JsonConvert.SerializeObject(shape, Utilities.GetDefaultJsonSerializerSettings())
        );
    }

    [Fact]
    public void ConvertNestedUnionMemberThroughSingleInnerUnion()
    {
        ImplicitConversionNested nested = new ImplicitConversionCircle { Kind = "", Radius = 1 };

        Assert.Equal(ImplicitConversionNestedType.ImplicitConversionShape, nested.Type);
        Assert.Equal(ImplicitConversionShapeType.Circle, nested.ImplicitConversionShape!.Type);
        Assert.Equal(1.0, nested.ImplicitConversionShape.ImplicitConversionCircle!.Radius);
        Assert.Equal(
            JsonConvert.SerializeObject(
                ImplicitConversionNested.CreateImplicitConversionShape(
                    ImplicitConversionShape.CreateCircle(
                        new ImplicitConversionCircle { Kind = "", Radius = 1 }
                    )
                ), Utilities.GetDefaultJsonSerializerSettings()),
            JsonConvert.SerializeObject(nested, Utilities.GetDefaultJsonSerializerSettings())
        );
    }

    [Fact]
    public void PreferDirectMemberOverNestedUnionMember()
    {
        ImplicitConversionNested nested = new ImplicitConversionSquare { Kind = "", Side = 3 };

        Assert.Equal(ImplicitConversionNestedType.ImplicitConversionSquare, nested.Type);
        Assert.NotNull(nested.ImplicitConversionSquare);
        Assert.Null(nested.ImplicitConversionShape);
    }

    [Fact]
    public void ConvertNullToNull()
    {
        ImplicitConversionCircle? circle = null;

        ImplicitConversionShape? shape = circle;
        ImplicitConversionNested? nested = circle;
        ImplicitConversionPrimitives? primitives = (string?)null;

        Assert.Null(shape);
        Assert.Null(nested);
        Assert.Null(primitives);
    }

    [Fact]
    public void ConvertNullableUnionMembers()
    {
        ImplicitConversionNullable value = new ImplicitConversionSquare { Kind = "", Side = 1 };

        Assert.Equal(ImplicitConversionNullableType.ImplicitConversionSquare, value.Type);
        Assert.NotNull(value.ImplicitConversionSquare);
    }

    [Fact]
    public void ConvertPrimitiveAndCollectionMembers()
    {
        ImplicitConversionPrimitives fromString = "text";
        ImplicitConversionPrimitives fromInt = 5;
        ImplicitConversionPrimitives fromDouble = 2.5;
        ImplicitConversionPrimitives fromList = new List<ImplicitConversionCircle>
        {
            new ImplicitConversionCircle { Kind = "circle", Radius = 1 },
        };
        ImplicitConversionPrimitives fromMap = new Dictionary<string, object> { { "key", "value" } };

        Assert.Equal("text", fromString.Str);
        Assert.Equal(5L, fromInt.Integer);
        Assert.Null(fromInt.Number);
        Assert.Equal(2.5, fromDouble.Number);
        Assert.Single(fromList.ArrayOfImplicitConversionCircle!);
        Assert.Equal("value", fromMap.MapOfAny!["key"]);
        Assert.Equal("5", JsonConvert.SerializeObject(fromInt, Utilities.GetDefaultJsonSerializerSettings()));
        Assert.Equal("\"text\"", JsonConvert.SerializeObject(fromString, Utilities.GetDefaultJsonSerializerSettings()));
    }

    [Fact]
    public void SkipMembersWithoutUnambiguousConversion()
    {
        Assert.Equal(
            new[]
            {
                nameof(ImplicitConversionCircle),
                nameof(ImplicitConversionPolygon),
                nameof(ImplicitConversionShape),
                nameof(ImplicitConversionSquare),
            },
            ImplicitConversionSources<ImplicitConversionNested>()
        );
        Assert.Equal(
            new[] { nameof(ImplicitConversionSquare) },
            ImplicitConversionSources<ImplicitConversionAliasedShape>()
        );
        Assert.Equal(
            new[] { nameof(String) },
            ImplicitConversionSources<ImplicitConversionWithAny>()
        );
    }

    [Fact]
    public void SkipCollectionsDifferingOnlyInItemNullability()
    {
        Assert.Equal(
            new[]
            {
                nameof(Boolean),
                nameof(ImplicitConversionNullableStringListOrBoolean),
                nameof(ImplicitConversionStringListOrInteger),
                nameof(Int64),
            },
            ImplicitConversionSources<ImplicitConversionNullableItems>()
        );
    }

    [Fact]
    public void SkipSourcesReachableThroughSeveralNestedUnions()
    {
        Assert.Equal(
            new[]
            {
                nameof(Boolean),
                nameof(Double),
                nameof(ImplicitConversionDeepInner),
                nameof(ImplicitConversionStringOrBoolean),
                nameof(ImplicitConversionStringOrInteger),
                nameof(ImplicitConversionStringOrNumber),
                nameof(Int64),
            },
            ImplicitConversionSources<ImplicitConversionDeepNested>()
        );
    }

    [Fact]
    public void ConvertUniqueSourceSeveralLevelsDeep()
    {
        ImplicitConversionDeepNested nested = true;

        Assert.Equal(ImplicitConversionDeepNestedType.ImplicitConversionDeepInner, nested.Type);
        Assert.Equal(
            ImplicitConversionDeepInnerType.ImplicitConversionStringOrBoolean,
            nested.ImplicitConversionDeepInner!.Type
        );
        Assert.True(nested.ImplicitConversionDeepInner.ImplicitConversionStringOrBoolean!.Boolean);
        Assert.Equal(
            JsonConvert.SerializeObject(
                ImplicitConversionDeepNested.CreateImplicitConversionDeepInner(
                    ImplicitConversionDeepInner.CreateImplicitConversionStringOrBoolean(
                        ImplicitConversionStringOrBoolean.CreateBoolean(true)
                    )
                ), Utilities.GetDefaultJsonSerializerSettings()),
            JsonConvert.SerializeObject(nested, Utilities.GetDefaultJsonSerializerSettings())
        );
    }
}
