export type ParameterType = 'string' | 'number' | 'boolean' | 'choice' | 'multi_choice';
export type ParameterEvent = 'manual' | 'deployment';

export interface Parameter {
  id: number;
  repo_id: number;
  // 同名变量可按入口分别定义；空值兼容为 manual。
  event: ParameterEvent;
  name: string;
  type: ParameterType;
  description: string;
  // 默认值只用于缺失字段；显式取消全部勾选不会回填默认值。
  default: string;
  options: string[];
  required: boolean;
  order: number;
  source: string;
}
