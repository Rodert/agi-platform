import { Select, Switch } from 'antd'
import type { AIModel } from '../../../types'

export const modelParamDefaults = (model?: AIModel) => Object.fromEntries(
  Object.entries(model?.params_config || {}).map(([key, config]) => [
    key,
    config.default ?? (config.type === 'switch' ? false : config.options?.[0]?.value),
  ])
)

export const modelParamCost = (model: AIModel | undefined, values: Record<string, unknown>) => {
  if (!model) return 0
  return Object.entries(model.params_config || {}).reduce((cost, [key, config]) => {
    const option = config.options?.find(item => item.value === values[key])
    return cost + (option?.extra_cost || 0)
  }, model.cost)
}

export function ModelParams({ model, values, onChange, compact = false, excludeKeys = [] }: {
  model?: AIModel
  values: Record<string, unknown>
  onChange: (values: Record<string, unknown>) => void
  compact?: boolean
  excludeKeys?: string[]
}) {
  const entries = Object.entries(model?.params_config || {}).filter(([key]) => !excludeKeys.includes(key))
  if (!entries.length) return null
  return <>{entries.map(([key, config]) => config.type === 'switch'
    ? <label key={key} className={compact ? 'theme-muted flex items-center gap-2 px-2 text-xs' : 'theme-muted mb-5 flex items-center justify-between'}><span>{config.label}</span><Switch size="small" checked={Boolean(values[key])} onChange={checked => onChange({ ...values, [key]: checked })}/></label>
    : config.options?.length ? <label key={key} className={compact ? '' : 'theme-muted mb-5 block'}>{!compact && <span className="mb-2 block">{config.label}</span>}<Select aria-label={config.label} variant={compact ? 'borderless' : 'outlined'} value={values[key] as string | undefined} onChange={value => onChange({ ...values, [key]: value })} options={config.options.map(option => ({ label: option.label, value: option.value }))} className={compact ? 'min-w-[92px]' : 'w-full'}/></label> : null
  )}</>
}
