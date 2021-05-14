import SSHKeysTable, { ISSHKeysTableProps } from './SSHKeysTable';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/AccountPage/SSHKeysTable',
  component: SSHKeysTable,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<ISSHKeysTableProps> = {
  sshKeys: [
    {
      name: 'Linux key',
      key:
        'AAAAB3NzaC1yc2EAAAADAQABAAACAQDc5qT5WnTtBE9wemalsVqec7YCuUF/6R6/RBhX+5j3kSiy4hSfu2Sk9QuE5CuDAIlZchgAManOXi+TO1OLvNietYfkDLZWRjcNFCyxxOQTkEsM0sU5B0zHtNfC0qjrrx5yzBQXpZSuAPsfk+awSaTO94GL6Cn3YV0qycNoHVaL91n+0B0lz5zy6yDoRBIZescPHKgxt3gCrv7i27BLJwLZhqxA9BaTMY75NXcSsxu0G/7UfnlqFR/kefCaU3ILdAkB4kWVnbdWKCQ9kcPTGZWPBXBZf5jfiAZJkWXDLi/FLSg1TaEae4M8mRYcpNI0OA3qGhd1321HckvLg45V6uL2jY9T5wfTJcYr/IdvRKROshw+YjmoDeYipYc1TPoJ7daLQT5fPqZps98+ftWTM5iIJ8W5fpa16VHUYLEVsKzzd2e6l+SG0u7j3gKSWz55n2QARc9VDHRR3er63i5jLrMPXxZKnS3hATdxXRRqR3TTLXF6yQ2UyoMg398gaDaqo4Ns6nOVU7K0jUhxD//g3jq5IODzTEPc66b93mXQkkCkRIKnfyAsULHLGebWMvy4ZsHtYMPJPjbqbWBmzEoQGFJIazPEFMzZUMjzfU/4z6KWLr+3/BguQteDp0wajjRhPGzV4NrKRq9WMFtwEk8jYsCOUhAxDMKK17lVQ5ZGrRjcVQ',
    },
    {
      name: 'MacOS key',
      key:
        'AAAAB3NzaC1yc2EAAAADAQABAAACAQDG5VtH8d0grbGlKmOSWwpef8bJ/WK6chtaHrQzcE1vnxeibIg9KyZf6i0pmDUhGaYgYO1I2RvixSjukBhx1Z1v4FEnYPCyFRZPrT3fMc8yOOwTwj08khctoZDVlx13VY3Yhto1TCS2qA2Tnl8vtn0fOgLRgTKAeZdCglLy39wXM0yuS1abhVYzzRwaKOM8NqyzerV2lA296YyudLpzXygPZgnkAP/SX0aaVuMXYDkc9RYne0LSBXtLwnI18veLB5raq4j6bEx/FBoXYTItyPeU9X6aDhcXx/DPdQRPpUMHrrYoBLeXwb/2tIBmsi/Bq31kjQ+yvt+pxnFG3KYydAFaB4vYmxo34R9ggW6wT5D64ZbhfDKE3ekXhYqhByG1Arni+o+I4oGWuf6qT/rdVZ7vGq77zrw3bhn2OUmTfG+wcEJqF+OfffZS2bNXjpouFGqVqA6BR30fuolu5pKBcyrVTcQt9R5Z5hS8uu3agnHVNedbIP9sF8RbczAPP4X3SKV9GdI0SZVwVTwVHTDYrByHJfogeVOv8WtPEwdSCGuTdKsQLzvrTj01P8Q91FnSLKR+Ks4uXl9K2+bVgX3xVmoLUg1ED2xZvJzYv/t481RipBTdM/vBRhkX86dlT2aFLlb3bxBC8jEhEFCKt9SMUgW/88Yejd3TA3yiwzjxJS89cw',
    },
    {
      name: 'Windows key',
      key:
        'AAAAB3NzaC1yc2EAAAADAQABAAACAQDkX4K/BQzL5+odDU3rdp6tZBwA3KsGP+eWVuOLLmbqt97Q1qL3AM3+cT1boKTlb8enSRVWMVDCCdECXXJx9HTXePQ/W2UtW5NhKxRxVgjBk5fYdESCl9T+076SrZ+oX0q62DggLb+yR0JQf/SA/7tHvtkdn4ztOhBd3ztjQ2lTEQn1XP1M03eSB1o8Vdro3BnRgBg/aJIK7sdPU+RyY4qso/SHOf/lkf394Pf+BH7VkesqBYkrz4xnSDYJRsJLT1U8YR0CZs6e3TJjazcywNKnTDeXoUL7xPRqBiTBJqbrKlglfhT9hK/b2wflEJM6eoGTWEL2spO2/h/JIafM3gJwGWSXG3wA2DPgcsA+7WL0PRYo48lqSGpjMHiBjXaZVi0Nj9jym7qGrHPQja/tkwtMgcl2RydtFtG8slpAVrQbBeLPvoAHFiID8yuag2WMoOAV5E0dWVURdB77vyaalExSimV2jj+kt2oMTo27zZB4Kbs4+ms1wFV7fR2GGaC4QJn+vP2ahBVM4fGXXCOJHEG/v5MnS/CYbPISJsOUMTxFbcsU475rg9XS+hCarPPyrqrix3M/vm/TRmFotz1CWknhA6LIHXgVTh0/8WdpK0D9urFyOl92G08U+wZfruZTqxyHbxu/YIUUmBUzrIMVA47O3Sq50mYeXFmyNw4f28NWVQ',
    },
  ],
};

const Template: Story<ISSHKeysTableProps> = args => <SSHKeysTable {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
