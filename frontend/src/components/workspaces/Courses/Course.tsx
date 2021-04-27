import { FC } from 'react';
import { Row, Col } from 'antd';
import { ApartmentOutlined } from '@ant-design/icons';
import '../Dashboard/Dashboard.css';
import '../../../index.less'; //To delete, usefull only to storybook

export interface ICourseProps {
  id: number;
  title: string;
  selected: boolean;
  onClick: (id: number) => void;
}

//border-solid border-4 rounded-3xl border-blue-500

const Course: FC<ICourseProps> = ({ ...props }) => {
  const { id, title, selected, onClick } = props;
  return (
    <Row className="sm:px-0 md:px-4">
      <Col span={24} className="flex justify-center pb-2">
        {/*<Badge
          count={5}
          offset={[-10, 10]}
          style={{ backgroundColor: '#1c7afd' }}
        ></Badge>*/}
        <button
          className={`row shadow-lg h-24 w-24 2xl:h-28 2xl:w-28 course ${
            selected ? ' active' : ''
          }`}
          onClick={() => onClick(id)}
        >
          <ApartmentOutlined style={{ fontSize: '30pt', color: 'black' }} />
        </button>
      </Col>
      <Col span={24} className="flex justify-center pb-0">
        <p className="w-28 h-8 2xl:h-12 2xl:w-32 text-center text-xs 2xl:text-sm pb-0">
          {title}
        </p>
      </Col>
    </Row>
  );
};

export default Course;
