import { FC } from 'react';
import Course from './Course';

export interface ICoursesProps {
  courses: Array<{ id: number; title: string }>;
  selected: number;
  onClick: (id: number) => void;
}

const Courses: FC<ICoursesProps> = ({ ...props }) => {
  const { courses, selected, onClick } = props;
  return (
    <div className="grid lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 lg:grid-flow-row grid-rows-1 grid-flow-col gap-3 lg:gap-4">
      {courses.map(course => (
        <div key={course.id} className="h-full flex justify-center">
          <Course
            id={course.id}
            title={course.title}
            selected={selected === course.id ? true : false}
            onClick={onClick}
          />
        </div>
      ))}
    </div>
  );
};

export default Courses;
