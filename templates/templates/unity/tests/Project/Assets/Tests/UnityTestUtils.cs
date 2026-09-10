using System;
using System.Collections;
using System.Threading.Tasks;

public static class UnityTestUtils
{
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
}
