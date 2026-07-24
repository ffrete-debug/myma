"use client";

import { useState } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { getAllPresets, getPresetsByType, presetToIni, ServerPreset } from '@/config/presets';

interface PresetSelectorProps {
  onSelect: (iniContent: string, presetName: string) => void;
}

export function PresetSelector({ onSelect }: PresetSelectorProps) {
  const [type, setType] = useState<'PVE' | 'PVP'>('PVE');
  const presets = getPresetsByType(type);

  return (
    <div className="space-y-3">
      <div className="flex gap-2">
        <Button variant={type === 'PVE' ? 'default' : 'outline'} size="sm" onClick={() => setType('PVE')}>PVE</Button>
        <Button variant={type === 'PVP' ? 'default' : 'outline'} size="sm" onClick={() => setType('PVP')}>PVP</Button>
      </div>
      <div className="space-y-2">
        {presets.map((p) => (
          <Card key={p.name} className="cursor-pointer hover:shadow-md transition-shadow" onClick={() => onSelect(presetToIni(p), p.label)}>
            <CardContent className="p-3">
              <div className="flex items-center gap-2 mb-1">
                <Badge variant={type === 'PVE' ? 'default' : 'destructive'} className="text-[10px] px-1.5 py-0">{p.multiplier}</Badge>
                <span className="font-semibold text-sm">{p.label}</span>
              </div>
              <p className="text-xs text-muted-foreground mb-1">{p.description}</p>
              <div className="flex flex-wrap gap-1">
                {p.changes.slice(0, 4).map((c) => (
                  <code key={c} className="text-[10px] bg-muted px-1 py-0.5 rounded">{c}</code>
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
