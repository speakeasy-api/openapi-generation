using NUnit.Framework;
using UnityEngine.TestTools;
using Openapi;
using System.Collections;

public class MultiLevelShould
{
    [UnityTest]
    public IEnumerator MultiLevelTest()
    {
        CommonHelpers.RecordTest("multi-level-grouping");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Nested.First.GetAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }
}
