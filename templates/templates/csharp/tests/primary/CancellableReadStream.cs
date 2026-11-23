#nullable enable
using System;
using System.IO;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

internal sealed class CancellableReadStream : MemoryStream
{
    public CancellableReadStream(string contents) : base(Encoding.UTF8.GetBytes(contents))
    {
    }

    public override ValueTask<int> ReadAsync(Memory<byte> buffer, CancellationToken cancellationToken = default)
    {
        return Position < Length
            ? base.ReadAsync(buffer, cancellationToken)
            : new ValueTask<int>(WaitForCancellation(cancellationToken));
    }

    public override Task<int> ReadAsync(byte[] buffer, int offset, int count, CancellationToken cancellationToken)
    {
        return Position < Length
            ? base.ReadAsync(buffer, offset, count, cancellationToken)
            : WaitForCancellation(cancellationToken);
    }

    private static async Task<int> WaitForCancellation(CancellationToken cancellationToken)
    {
        await Task.Delay(Timeout.Infinite, cancellationToken);
        return 0;
    }
}
