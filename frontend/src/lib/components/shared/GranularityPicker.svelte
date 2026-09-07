<script lang="ts">
  import { SegmentedControl } from "@kenn-io/kit-ui";
  import type { Granularity } from "../../stores/analytics.svelte.js";
  import { m } from "../../i18n/index.js";

  interface Props {
    value: Granularity;
    onChange: (value: Granularity) => void;
    disabled?: boolean;
  }

  let { value, onChange, disabled = false }: Props = $props();

  const options = $derived([
    { value: "day", label: m.analytics_granularity_day() },
    { value: "week", label: m.analytics_granularity_week() },
    { value: "month", label: m.analytics_granularity_month() },
  ]);
</script>

<SegmentedControl
  {options}
  {value}
  {disabled}
  ariaLabel={m.analytics_granularity_label()}
  onchange={(v) => onChange(v as Granularity)}
/>
