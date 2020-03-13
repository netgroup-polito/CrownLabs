import React, { Component } from 'react';
import PropTypes from 'prop-types';
import WebFont from 'webfontloader';

import './index.css';
import { sizes, light, dark, mdInactive } from './config/mappings';

class MaterialIcon extends Component {
    constructor(props) {
        super(props);

        const {preloader} = this.props;
        
        this.state = {
            element: preloader
        }

        this.onFontActive = this.onFontActive.bind(this);
        this.processProps = this.processProps.bind(this);

        WebFont.load({
            google: {
                families: ['Material+Icons']
            },
            timeout:5000,
            fontactive: this.onFontActive
        })
    }

    componentDidMount() {

    }

    onFontActive(fontFamily, fvd) {
        const {icon, styleOverride, clsName, ...other} = this.processProps();
        this.setState({element: <i {...other} className={clsName} style={styleOverride} >{icon}</i>})
    }

    processProps() {
        const {size, invert, inactive, style, className, color, icon, ...other} = this.props;

        const sizeMapped = sizes[size] || parseInt(size) || sizes['small'];
        const defaultColor = (invert && 'invert') ? light : dark;
        const inactiveColor = (inactive && 'inactive') ? mdInactive : '';
        const propStyle = style || {};
        const styleOverride = Object.assign(propStyle, {color: color ? color : '', fontSize: sizeMapped});
        const clsName = className || `material-icons ${sizeMapped} ${defaultColor} ${inactiveColor}`;

        return {
            icon, styleOverride, clsName, ...other
        }
    }

    render() {
        const {styleOverride, clsName, ...other} = this.processProps();

        return (this.state.element || <i {...other} className={clsName} style={styleOverride} ></i>)
    }
}

MaterialIcon.propTypes = {
    icon: PropTypes.string.isRequired,
    size: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    invert: PropTypes.bool,
    inactive: PropTypes.bool
};

export default MaterialIcon;
export * from './palette';
