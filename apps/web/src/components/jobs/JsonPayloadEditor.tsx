import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { WandSparkles } from "lucide-react";

interface JsonPayloadEditorProps {
  value: string;
  onChange: (value: string) => void;
  error?: string;
}

export function JsonPayloadEditor({
  value,
  onChange,
  error,
}: JsonPayloadEditorProps) {
  function formatJSON() {
    try {
      const parsed = JSON.parse(value);
      onChange(JSON.stringify(parsed, null, 2));
    } catch {
      // The schema validation will display the error.
    }
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium">JSON Payload</label>

        <Button type="button" variant="outline" size="sm" onClick={formatJSON}>
          <WandSparkles className="mr-2 size-4" />
          Format JSON
        </Button>
      </div>

      <Textarea
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={'{\n  "email": "developer@example.com"\n}'}
        className="min-h-52 resize-y font-mono text-sm"
      />

      {error && <p className="text-sm text-red-400">{error}</p>}
    </div>
  );
}
