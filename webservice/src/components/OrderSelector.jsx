import React from 'react';
import IconButton from '@material-ui/core/IconButton';
import makeStyles from '@material-ui/core/styles/makeStyles';
import ArrowUpwardIcon from '@material-ui/icons/ArrowUpward';
import Selector from './Selector';

const useStyles = makeStyles(() => ({
  orderSelector: { display: 'flex' },
  flipOrderBtn: {
    transition: '0.5s ease-out'
  },
  rotate0: {
    transform: 'rotateZ(0deg)'
  },
  rotate180: {
    transform: 'rotateZ(180deg)'
  }
}));

export default function OrderSelector(props) {
  const classes = useStyles();
  const { selectors, orderData, setOrderData } = props;
  const { order, isDirUp } = orderData;
  const setOrder = newOrder => {
    setOrderData({ ...orderData, order: newOrder });
  };
  return (
    <div className={classes.orderSelector}>
      <Selector selectors={selectors} value={order} setValue={setOrder} />
      <IconButton
        color="secondary"
        onClick={() => setOrderData({ ...orderData, isDirUp: !isDirUp })}
        className={`${classes.flipOrderBtn} ${
          isDirUp ? classes.rotate180 : classes.rotate0
        }`}
      >
        <ArrowUpwardIcon fontSize="large" />
      </IconButton>
    </div>
  );
}
