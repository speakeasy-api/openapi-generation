using System;
using System.Collections.Generic;
using System.Collections;
using System.IO;
using System.Threading.Tasks;
using NUnit.Framework;

public static class CommonHelpers
{
    // Test service URLs - read from environment with defaults for local development
    public static readonly string HttpBinPort = Environment.GetEnvironmentVariable("HTTPBIN_PORT") ?? "35123";
    public static readonly string ApiTestServicePort = Environment.GetEnvironmentVariable("API_TEST_SERVICE_PORT") ?? "35456";
    public static readonly string HttpBinUrl = $"http://localhost:{HttpBinPort}";
    public static readonly string ApiTestServiceUrl = $"http://localhost:{ApiTestServicePort}";

    private static Object obj = new Object();

    public static void RecordTest(string id)
    {
        lock (CommonHelpers.obj)
        {
            File.AppendAllLines("test-unity-record.txt", new List<string> { id });
        }
    }

    public static IEnumerator Await(Task task)
    {
        while (!task.IsCompleted)
        {
            yield return null;
        }
        if (task.IsFaulted)
        {
            throw task.Exception;
        }
    }

    public static IEnumerator Await(Func<Task> taskDelegate)
    {
        return Await(taskDelegate.Invoke());
    }
    public static void AssertObjectEquivalent<T>(T a, T o)
    {
        if (a == null && o == null) {
            return;
        }
        Assert.NotNull(a);
        Assert.NotNull(o);
        Assert.AreEqual(a.GetType(), o.GetType());
        var isArray = typeof(IList).IsAssignableFrom(a.GetType());
        var isSimple = (a.GetType() is not object) || (a.GetType() == typeof(string)) || (a.GetType() == typeof(String)) || (a.GetType().IsValueType);
        var isDict =  typeof(IDictionary).IsAssignableFrom(a.GetType());
        if (isSimple) {
            Assert.AreEqual(a, o);
            return;
        } else if (isArray) {
            AssertArraysEqual((IList)a, (IList)o);
            return;
        } else if (isDict) {
            AssertDictsEqual((IDictionary)a, (IDictionary)o);
            return;
        }
        var properties = a.GetType().GetProperties();
        foreach(var prop in properties) {
            if ((prop.GetMethod?.IsStatic ?? false) || (prop.SetMethod?.IsStatic ?? false)) {continue; }
            var propertyType = prop.PropertyType;

            var propA = prop.GetValue(a);
            var propO = prop.GetValue(o);
            if (propA == null && propO == null) {
                continue;
            }
            AssertObjectEquivalent(propA, propO);
        }
    }
    public static void AssertDictsEqual(IDictionary dict1, IDictionary dict2) {
        if (dict1.Count != dict2.Count) {
            Assert.Fail("Dicts have different lengths");
        }
        foreach (var key in dict1.Keys) {
            if (!dict2.Contains(key)) {
                Assert.Fail("Dicts have different keys");
            }
            var val1 = dict1[key];
            var val2 = dict2[key];
            AssertObjectEquivalent(val1, val2);
        }
    }
    public static void AssertArraysEqual(IList arr1, IList arr2) {
        int index = 0;
        foreach (var item in arr1) {
            var item2 = arr2[index];
            AssertObjectEquivalent(item, item2);
            index++;
        }
    }

    public struct TestTableEntry {
        public string name;
        public object arg;
        public object want;
        public string testId;
    }
}
