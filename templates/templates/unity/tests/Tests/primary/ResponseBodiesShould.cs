using NUnit.Framework;
using System;
using System.Collections;
using System.Collections.Generic;
using System.Numerics;
using UnityEngine.TestTools;
using Openapi;
using Openapi.Models.Shared;
using Openapi.Utils;

public class ResponseBodiesShould
{
    [UnityTest]
    public IEnumerator JsonGet()
    {
        CommonHelpers.RecordTest("response-bodies-json-get");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.ResponseBodyJsonGetAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.HttpBinSimpleJsonObject);

                var slideshow = res.HttpBinSimpleJsonObject.Slideshow;
                Assert.AreEqual("Yours Truly", slideshow.Author);
                Assert.AreEqual("date of publication", slideshow.Date);
                Assert.AreEqual("Sample Slide Show", slideshow.Title);

                var slides = slideshow.Slides.ToArray();
                Assert.AreEqual("Wake up to WonderWidgets!", slides[0].Title);
                Assert.AreEqual("all", slides[0].Type);
                Assert.AreEqual("Overview", slides[1].Title);
                Assert.AreEqual("all", slides[1].Type);

                var items = slides[1].Items.ToArray();
                Assert.AreEqual("Why <em>WonderWidgets</em> are great", items[0]);
                Assert.AreEqual("Who <em>buys</em> WonderWidgets", items[1]);
            }
        });
    }

    [UnityTest]
    public IEnumerator StringGet()
    {
        CommonHelpers.RecordTest("response-bodies-string-get");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.ResponseBodies.ResponseBodyStringGetAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.False(string.IsNullOrEmpty(res.Html));
                Assert.AreEqual(
                    "<!DOCTYPE html>\n<html>\n  <head>\n  </head>\n  <body>\n      <h1>Herman Melville - Moby-Dick</h1>\n\n      <div>\n        <p>\n          Availing himself of the mild, summer-cool weather that now reigned in these latitudes, and in preparation for the peculiarly active pursuits shortly to be anticipated, Perth, the begrimed, blistered old blacksmith, had not removed his portable forge to the hold again, after concluding his contributory work for Ahab's leg, but still retained it on deck, fast lashed to ringbolts by the foremast; being now almost incessantly invoked by the headsmen, and harpooneers, and bowsmen to do some little job for them; altering, or repairing, or new shaping their various weapons and boat furniture. Often he would be surrounded by an eager circle, all waiting to be served; holding boat-spades, pike-heads, harpoons, and lances, and jealously watching his every sooty movement, as he toiled. Nevertheless, this old man's was a patient hammer wielded by a patient arm. No murmur, no impatience, no petulance did come from him. Silent, slow, and solemn; bowing over still further his chronically broken back, he toiled away, as if toil were life itself, and the heavy beating of his hammer the heavy beating of his heart. And so it was.—Most miserable! A peculiar walk in this old man, a certain slight but painful appearing yawing in his gait, had at an early period of the voyage excited the curiosity of the mariners. And to the importunity of their persisted questionings he had finally given in; and so it came to pass that every one now knew the shameful story of his wretched fate. Belated, and not innocently, one bitter winter's midnight, on the road running between two country towns, the blacksmith half-stupidly felt the deadly numbness stealing over him, and sought refuge in a leaning, dilapidated barn. The issue was, the loss of the extremities of both feet. Out of this revelation, part by part, at last came out the four acts of the gladness, and the one long, and as yet uncatastrophied fifth act of the grief of his life's drama. He was an old man, who, at the age of nearly sixty, had postponedly encountered that thing in sorrow's technicals called ruin. He had been an artisan of famed excellence, and with plenty to do; owned a house and garden; embraced a youthful, daughter-like, loving wife, and three blithe, ruddy children; every Sunday went to a cheerful-looking church, planted in a grove. But one night, under cover of darkness, and further concealed in a most cunning disguisement, a desperate burglar slid into his happy home, and robbed them all of everything. And darker yet to tell, the blacksmith himself did ignorantly conduct this burglar into his family's heart. It was the Bottle Conjuror! Upon the opening of that fatal cork, forth flew the fiend, and shrivelled up his home. Now, for prudent, most wise, and economic reasons, the blacksmith's shop was in the basement of his dwelling, but with a separate entrance to it; so that always had the young and loving healthy wife listened with no unhappy nervousness, but with vigorous pleasure, to the stout ringing of her young-armed old husband's hammer; whose reverberations, muffled by passing through the floors and walls, came up to her, not unsweetly, in her nursery; and so, to stout Labor's iron lullaby, the blacksmith's infants were rocked to slumber. Oh, woe on woe! Oh, Death, why canst thou not sometimes be timely? Hadst thou taken this old blacksmith to thyself ere his full ruin came upon him, then had the young widow had a delicious grief, and her orphans a truly venerable, legendary sire to dream of in their after years; and all of them a care-killing competency.\n        </p>\n      </div>\n  </body>\n</html>",
                    res.Html
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator XmlGet()
    {
        CommonHelpers.RecordTest("response-bodies-xml-get");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.ResponseBodies.ResponseBodyXmlGetAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "<?xml version='1.0' encoding='us-ascii'?>\n\n<!--  A SAMPLE set of slides  -->\n\n<slideshow \n    title=\"Sample Slide Show\"\n    date=\"Date of publication\"\n    author=\"Yours Truly\"\n    >\n\n    <!-- TITLE SLIDE -->\n    <slide type=\"all\">\n      <title>Wake up to WonderWidgets!</title>\n    </slide>\n\n    <!-- OVERVIEW -->\n    <slide type=\"all\">\n        <title>Overview</title>\n        <item>Why <em>WonderWidgets</em> are great</item>\n        <item/>\n        <item>Who <em>buys</em> WonderWidgets</item>\n    </slide>\n\n</slideshow>",
                    res.Xml
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator BytesGet()
    {
        CommonHelpers.RecordTest("response-bodies-bytes-get");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.ResponseBodies.ResponseBodyBytesGetAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Bytes);
                Assert.AreEqual(100, res.Bytes.Length);
            }
        });
    }

    [UnityTest]
    public IEnumerator ResponseBodyReadOnly()
    {
        CommonHelpers.RecordTest("response-bodies-read-only");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.ResponseBodies.ResponseBodyReadOnlyAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.True(res.ReadOnlyObject.Bool);
                Assert.AreEqual(1.0, res.ReadOnlyObject.Num);
                Assert.AreEqual("hello", res.ReadOnlyObject.String);
            }
        });
    }

    [UnityTest]
    public IEnumerator ResponseBodyAdditionalPropertiesString()
    {
        CommonHelpers.RecordTest("response-bodies-additional-properties-string");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new ObjWithStringAdditionalProperties()
            {
                NormalField = "string",
            };

            var json = Helpers.GetSerializedBodyJson(req);
            Assert.AreEqual("{\"normalField\":\"string\"}", json);

            using (var res = await sdk.ResponseBodies.ResponseBodyAdditionalPropertiesPostAsync(req))
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("string", res.Object.Json.NormalField);
                Assert.Null(res.Object.Json.AdditionalProperties);
            }

            req = new ObjWithStringAdditionalProperties()
            {
                NormalField = "string",
                AdditionalProperties = new Dictionary<string, string>()
                {
                    { "extra1", "value1" },
                    { "extra2", null },
                },
            };

            json = Helpers.GetSerializedBodyJson(req);
            Assert.AreEqual("{\"normalField\":\"string\",\"additionalProperties\":{\"extra1\":\"value1\",\"extra2\":null}}", json);
            using (var res = await sdk.ResponseBodies.ResponseBodyAdditionalPropertiesPostAsync(req))
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req.NormalField, res.Object.Json.NormalField);
                Assert.AreEqual(req.AdditionalProperties, res.Object.Json.AdditionalProperties);
            }
        });
    }

    [UnityTest]
    public IEnumerator ResponseBodyAdditionalPropertiesDate()
    {
        CommonHelpers.RecordTest("response-bodies-additional-properties-date");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new ObjWithDateAdditionalProperties()
            {
                NormalField = "string",
                AdditionalProperties = new Dictionary<string, DateOnly>()
                {
                    { "today", DateOnly.FromDateTime(new DateTime(2020, 1, 1)) }
                }
            };

            var json = Helpers.GetSerializedBodyJson(req);
            Assert.AreEqual("{\"normalField\":\"string\",\"additionalProperties\":{\"today\":\"2020-01-01\"}}", json);

            using (var res = await sdk.ResponseBodies.ResponseBodyAdditionalPropertiesDatePostAsync(req))
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req.NormalField, res.Object.Json.NormalField);
                Assert.AreEqual(req.AdditionalProperties, res.Object.Json.AdditionalProperties);
            }
        });
    }

    [UnityTest]
    public IEnumerator ResponseBodyAdditionalPropertiesComplexNumbers()
    {
        CommonHelpers.RecordTest("response-bodies-additional-properties-complex-numbers");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new ObjWithComplexNumbersAdditionalProperties()
            {
                NormalField = "string",
                AdditionalProperties = new Dictionary<string, BigInteger>()
                {
                    { "bigintStr", BigInteger.Parse("123456789012345678901234567890") }
                }
            };

            var json = Helpers.GetSerializedBodyJson(req);
            Assert.AreEqual("{\"normalField\":\"string\",\"additionalProperties\":{\"bigintStr\":\"123456789012345678901234567890\"}}", json);

            using (var res = await sdk.ResponseBodies.ResponseBodyAdditionalPropertiesComplexNumbersPostAsync(req))
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req.NormalField, res.Object.Json.NormalField);
                Assert.AreEqual(req.AdditionalProperties, res.Object.Json.AdditionalProperties);
            }
        });
    }

    [UnityTest]
    public IEnumerator ResponseBodyAdditionalPropertiesObject()
    {
        CommonHelpers.RecordTest("response-bodies-additional-properties-object");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var obj = Helpers.CreateSimpleObject();
            var req = new ObjWithObjAdditionalProperties()
            {
                Datetime = new System.DateTime(2020, 1, 1, 0, 0, 0, System.DateTimeKind.Utc),
                AdditionalProperties = new System.Collections.Generic.List<long> { 1, 2, 3 },
                AdditionalPropertiesT = new System.Collections.Generic.Dictionary<string, SimpleObject>()
                {
                    { "obj1", obj }
                }
            };

            var json = Helpers.GetSerializedBodyJson(req);
            Assert.IsTrue(json.Contains("{\"AdditionalProperties\":[1,2,3],\"datetime\":\"2020-01-01T00:00:00.0000000Z\",\"additionalProperties\":{\"obj1\":{\"any\":\"any\","));

            using (var res = await sdk.ResponseBodies.ResponseBodyAdditionalPropertiesObjectPostAsync(req))
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req.Datetime, res.Object.Json.Datetime);
                Assert.AreEqual(req.AdditionalProperties, res.Object.Json.AdditionalProperties);
                Helpers.AssertSimpleObject(res.Object.Json.AdditionalPropertiesT["obj1"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator ResponseBodyAdditionalPropertiesObjectAnyValues()
    {
        CommonHelpers.RecordTest("response-bodies-additional-properties-any-values");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new ObjWithAnyAdditionalProperties();

            using (var res = await sdk.ResponseBodies.ResponseBodyAdditionalPropertiesAnyPostAsync(req))
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.Null(res.Object.Json.NormalField);
                Assert.Null(res.Object.Json.AdditionalProperties);
            }

            req = new ObjWithAnyAdditionalProperties()
            {
                AdditionalProperties = new System.Collections.Generic.Dictionary<string, object>()
                {
                    { "key1", "value1" },
                    { "key2", null },
                    { "key3", new System.Collections.Generic.Dictionary<string, object>()
                        {
                            { "foo", "bar" },
                            { "subkey1", new System.Collections.Generic.Dictionary<string, object>()
                                {
                                    { "foo", "bar" },
                                    { "nan", null }
                                }
                            }
                        }
                    },
                    { "key4", new System.Collections.Generic.List<object>() { "foo", "bar" } }
                }
            };

            using (var res = await sdk.ResponseBodies.ResponseBodyAdditionalPropertiesAnyPostAsync(req))
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.Null(res.Object.Json.NormalField);
                Helpers.AssertDictEqual(req.AdditionalProperties, res.Object.Json.AdditionalProperties);
            }
        });
    }
}
