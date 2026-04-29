import { Form, InputNumber } from 'antd';
import type { FC, ReactNode } from 'react';
import type { Rule } from 'antd/es/form';

interface QuotaFieldRuleSet {
  cpu?: Rule[];
  memory?: Rule[];
  instances?: Rule[];
}

interface QuotaFieldLimit {
  min?: number;
  max?: number;
}

interface QuotaFieldLimitSet {
  cpu?: QuotaFieldLimit;
  memory?: QuotaFieldLimit;
  instances?: QuotaFieldLimit;
}

interface QuotaFieldTooltipSet {
  cpu?: ReactNode;
  memory?: ReactNode;
  instances?: ReactNode;
}

export interface IQuotaFieldsProps {
  disabled?: boolean;
  validateTrigger?: string;
  rules?: QuotaFieldRuleSet;
  limits?: QuotaFieldLimitSet;
  tooltips?: QuotaFieldTooltipSet;
}

const QuotaFields: FC<IQuotaFieldsProps> = ({
  disabled = false,
  validateTrigger,
  rules,
  limits,
  tooltips,
}) => {
  return (
    <>
      <Form.Item
        name="cpu"
        label="CPU"
        validateTrigger={validateTrigger}
        rules={rules?.cpu}
        tooltip={tooltips?.cpu}
      >
        <InputNumber
          min={limits?.cpu?.min}
          max={limits?.cpu?.max}
          disabled={disabled}
          className="w-100"
        />
      </Form.Item>

      <Form.Item
        name="memory"
        label="Memory (Gi)"
        validateTrigger={validateTrigger}
        rules={rules?.memory}
        tooltip={tooltips?.memory}
      >
        <InputNumber
          min={limits?.memory?.min}
          max={limits?.memory?.max}
          disabled={disabled}
          className="w-100"
          addonAfter="Gi"
        />
      </Form.Item>

      <Form.Item
        name="instances"
        label="Instances"
        validateTrigger={validateTrigger}
        rules={rules?.instances}
        tooltip={tooltips?.instances}
      >
        <InputNumber
          min={limits?.instances?.min}
          max={limits?.instances?.max}
          disabled={disabled}
          className="w-100"
        />
      </Form.Item>
    </>
  );
};

export default QuotaFields;
