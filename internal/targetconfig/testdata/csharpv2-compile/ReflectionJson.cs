namespace Acme.Widgets
{
    using System.Collections.Generic;
    using System.Text.Json;

    public class ReflectionJson
    {
        public ReflectionJson(string json)
        {
            Defaults = JsonSerializer.Deserialize<Dictionary<string, string>>(json) ?? new Dictionary<string, string>();
        }

        public Dictionary<string, string> Defaults { get; }

        public string Serialize(Dictionary<string, string> values)
        {
            return JsonSerializer.Serialize(values);
        }
    }
}
